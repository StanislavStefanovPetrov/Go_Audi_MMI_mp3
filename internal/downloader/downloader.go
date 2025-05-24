package downloader

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"io/ioutil"

	"github.com/bogem/id3v2"
	"github.com/stanislavpetrov/Go_Audi_MMI_mp3/internal/config"
)

// sanitizeMetadata removes non-ASCII characters and emojis from metadata fields
func sanitizeMetadata(text string) string {
	// Remove emojis and non-ASCII characters
	reg := regexp.MustCompile(`[^\x00-\x7F]`)
	text = reg.ReplaceAllString(text, "")

	// Remove multiple spaces and newlines
	reg = regexp.MustCompile(`\s+`)
	text = reg.ReplaceAllString(text, " ")

	// Trim spaces
	text = strings.TrimSpace(text)

	return text
}

// sanitizeFilename removes non-ASCII characters, emojis, and invalid filesystem characters
func sanitizeFilename(filename string) string {
	// Remove emojis and non-ASCII characters
	reg := regexp.MustCompile(`[^\x00-\x7F]`)
	filename = reg.ReplaceAllString(filename, "")

	// Replace invalid filesystem characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "'", "&", "+", "=", "!", "@", "#", "$", "%", "^", "(", ")", "[", "]", "{", "}", ";", ",", "`", "~"}
	result := filename
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Remove multiple consecutive underscores
	reg = regexp.MustCompile(`_+`)
	result = reg.ReplaceAllString(result, "_")

	// Trim spaces and underscores from start and end
	result = strings.Trim(result, " _")

	// Ensure the filename is not empty after sanitization
	if result == "" {
		result = "untitled"
	}

	return result
}

func removeNonASCII(text string) string {
	// Remove non-ASCII
	result := make([]rune, 0, len(text))
	for _, r := range text {
		if r <= 127 {
			result = append(result, r)
		}
	}
	ascii := string(result)
	// Remove non-alphanumeric except basic punctuation
	re := regexp.MustCompile(`[^a-zA-Z0-9\s.,!?()\-]+`)
	ascii = re.ReplaceAllString(ascii, "")
	// Replace whitespace with a single space
	re = regexp.MustCompile(`\s+`)
	ascii = re.ReplaceAllString(ascii, " ")
	return strings.TrimSpace(ascii)
}

type Downloader struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Downloader {
	return &Downloader{
		cfg: cfg,
	}
}

func (d *Downloader) Download(ctx context.Context, videoURL string) error {
	// Prepare output filename pattern with sanitized title
	outputPattern := filepath.Join(d.cfg.OutputDir, "%(title)s.%(ext)s")

	// Prepare yt-dlp command with sanitized metadata
	args := []string{
		"-x", // extract audio
		"--audio-format", "mp3",
		"--audio-quality", fmt.Sprintf("%dK", d.cfg.Bitrate),
		"--add-metadata",
		"--embed-metadata",
		// Use regex to clean metadata
		"--parse-metadata", "title:(?s)(.*)",
		"--parse-metadata", "artist:(?s)(.*)",
		"--parse-metadata", "description:(?s)(.*)",
		// Remove non-ASCII characters from metadata
		"--parse-metadata", "title:regex_replace:[^\\x00-\\x7F]:",
		"--parse-metadata", "artist:regex_replace:[^\\x00-\\x7F]:",
		"--parse-metadata", "description:regex_replace:[^\\x00-\\x7F]:",
		// Clean up multiple spaces
		"--parse-metadata", "title:regex_replace:\\s+: :",
		"--parse-metadata", "artist:regex_replace:\\s+: :",
		"--parse-metadata", "description:regex_replace:\\s+: :",
		// Set album as combination of artist and title
		"--parse-metadata", "album:%(artist)s - %(title)s",
		"--restrict-filenames",
		"--no-playlist",
		"--no-write-info-json",
		"--no-write-thumbnail",
		"--no-write-subs",
		"--no-write-auto-subs",
		"--clean-infojson",
		"-o", outputPattern,
		videoURL,
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %s, error: %w", string(output), err)
	}

	// Get the actual filename from yt-dlp output and sanitize it
	if strings.Contains(string(output), "Destination:") {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Destination:") {
				parts := strings.Split(line, "Destination:")
				if len(parts) > 1 {
					oldPath := strings.TrimSpace(parts[1])
					newPath := filepath.Join(
						filepath.Dir(oldPath),
						sanitizeFilename(filepath.Base(oldPath)),
					)
					if oldPath != newPath {
						exec.Command("mv", oldPath, newPath).Run()
					}
				}
			}
		}
	}

	// Find the output file (should be only one new mp3 in output dir)
	files, err := ioutil.ReadDir(d.cfg.OutputDir)
	if err != nil {
		return err
	}
	var mp3file string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".mp3") {
			mp3file = filepath.Join(d.cfg.OutputDir, f.Name())
		}
	}
	if mp3file == "" {
		return fmt.Errorf("no mp3 file found after download")
	}

	// Open and clean tags
	tag, err := id3v2.Open(mp3file, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open mp3 for tag cleaning: %w", err)
	}
	defer tag.Close()

	tag.SetTitle(removeNonASCII(tag.Title()))
	tag.SetArtist(removeNonASCII(tag.Artist()))
	tag.SetAlbum(removeNonASCII(tag.Album()))

	// Clean all CommentFrames (COMM)
	for _, cf := range tag.GetFrames(tag.CommonID("Comments")) {
		if comment, ok := cf.(id3v2.CommentFrame); ok {
			comment.Text = removeNonASCII(comment.Text)
			tag.DeleteFrames(tag.CommonID("Comments"))
			tag.AddCommentFrame(comment)
		}
	}

	if err := tag.Save(); err != nil {
		return fmt.Errorf("failed to save cleaned tags: %w", err)
	}

	// Clean description and synopsis using ffmpeg (overwrite with ASCII-only)
	ffmpegTmp := strings.TrimSuffix(mp3file, ".mp3") + ".tmp.mp3"
	ffmpegArgs := []string{"-y", "-i", mp3file, "-metadata", "description=", "-metadata", "synopsis=", "-c:a", "copy", ffmpegTmp}
	cmd = exec.Command("ffmpeg", ffmpegArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg metadata clean failed: %s, error: %w", string(out), err)
	}
	// Replace original file
	if err := exec.Command("mv", ffmpegTmp, mp3file).Run(); err != nil {
		return fmt.Errorf("failed to replace mp3 after ffmpeg metadata clean: %w", err)
	}

	fmt.Printf("Successfully downloaded, converted and cleaned tags: %s\n", videoURL)
	return nil
}

func (d *Downloader) DownloadBatch(ctx context.Context, urls []string) error {
	for _, url := range urls {
		if err := d.Download(ctx, url); err != nil {
			fmt.Printf("Error downloading %s: %v\n", url, err)
			continue
		}
	}
	return nil
}
