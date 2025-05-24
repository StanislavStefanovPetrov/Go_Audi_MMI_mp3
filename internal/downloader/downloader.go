package downloader

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stanislavpetrov/Go_Audi_MMI_mp3/internal/config"
)

type Downloader struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Downloader {
	return &Downloader{
		cfg: cfg,
	}
}

func (d *Downloader) Download(ctx context.Context, videoURL string) error {
	// Prepare output filename pattern (yt-dlp will replace %(title)s)
	outputPattern := filepath.Join(d.cfg.OutputDir, "%(title)s.%(ext)s")

	// Prepare yt-dlp command
	args := []string{
		"-x", // extract audio
		"--audio-format", "mp3",
		"--audio-quality", fmt.Sprintf("%dK", d.cfg.Bitrate),
		"-o", outputPattern,
		videoURL,
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %s, error: %w", string(output), err)
	}

	fmt.Printf("Successfully downloaded and converted: %s\n", videoURL)
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

func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
