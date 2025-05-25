package downloader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"io/ioutil"

	"github.com/bogem/id3v2"
	"github.com/stanislavpetrov/Go_Audi_MMI_mp3/internal/config"
)

var debug bool

func init() {
	debugEnv := os.Getenv("DEBUG")
	debug = (debugEnv == "true" || debugEnv == "1")
	if debug {
		fmt.Println("[DEBUG] Debug mode enabled")
	}
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
	if debug {
		fmt.Printf("[DEBUG] Starting download for URL: %s\n", videoURL)
		fmt.Printf("[DEBUG] Config: OutputDir=%s, Bitrate=%d, SampleRate=%d, Channels=%d\n",
			d.cfg.OutputDir, d.cfg.Bitrate, d.cfg.SampleRate, d.cfg.Channels)
	}

	// Prepare output filename pattern with sanitized title
	outputPattern := filepath.Join(d.cfg.OutputDir, "%(title)s.%(ext)s")
	if debug {
		fmt.Printf("[DEBUG] Output pattern: %s\n", outputPattern)
	}

	// Prepare yt-dlp command with minimal required arguments
	args := []string{
		"-x", // extract audio
		"--audio-format", "mp3",
		"--restrict-filenames",
		"--no-playlist",
		"-o", outputPattern,
		videoURL,
	}

	if debug {
		fmt.Printf("[DEBUG] yt-dlp command: yt-dlp %v\n", args)
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if debug {
			fmt.Printf("[DEBUG] yt-dlp failed with output: %s\n", string(output))
		}
		return fmt.Errorf("yt-dlp failed: %s, error: %w", string(output), err)
	}

	if debug {
		fmt.Printf("[DEBUG] yt-dlp output: %s\n", string(output))
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
			if debug {
				fmt.Printf("[DEBUG] Found MP3 file: %s\n", mp3file)
			}
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
	if debug {
		fmt.Printf("[DEBUG] ffmpeg metadata clean command: ffmpeg %v\n", ffmpegArgs)
	}

	cmd = exec.Command("ffmpeg", ffmpegArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		if debug {
			fmt.Printf("[DEBUG] ffmpeg metadata clean failed with output: %s\n", string(out))
		}
		return fmt.Errorf("ffmpeg metadata clean failed: %s, error: %w", string(out), err)
	}

	// Replace original file
	if debug {
		fmt.Printf("[DEBUG] Moving %s to %s\n", ffmpegTmp, mp3file)
	}
	if err := exec.Command("mv", ffmpegTmp, mp3file).Run(); err != nil {
		return fmt.Errorf("failed to replace mp3 after ffmpeg metadata clean: %w", err)
	}

	// FINAL: Ensure correct sample rate, channels, bitrate with ffmpeg
	ffmpegFinal := strings.TrimSuffix(mp3file, ".mp3") + ".final.mp3"
	ffmpegArgs = []string{
		"-y",
		"-i", mp3file,
		"-ar", fmt.Sprintf("%d", d.cfg.SampleRate),
		"-ac", fmt.Sprintf("%d", d.cfg.Channels),
		"-b:a", fmt.Sprintf("%dk", d.cfg.Bitrate),
		"-c:a", "libmp3lame",
		ffmpegFinal,
	}
	if debug {
		fmt.Printf("[DEBUG] ffmpeg final conversion command: ffmpeg %v\n", ffmpegArgs)
	}

	cmd = exec.Command("ffmpeg", ffmpegArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		if debug {
			fmt.Printf("[DEBUG] ffmpeg final conversion failed with output: %s\n", string(out))
		}
		return fmt.Errorf("ffmpeg final sample rate conversion failed: %s, error: %w", string(out), err)
	}

	// Replace original file with the one that has correct sample rate
	if debug {
		fmt.Printf("[DEBUG] Moving %s to %s\n", ffmpegFinal, mp3file)
	}
	if err := exec.Command("mv", ffmpegFinal, mp3file).Run(); err != nil {
		return fmt.Errorf("failed to replace mp3 after ffmpeg final sample rate conversion: %w", err)
	}

	if debug {
		fmt.Printf("[DEBUG] Download completed successfully for: %s\n", videoURL)
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
