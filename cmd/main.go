package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stanislavpetrov/Go_Audi_MMI_mp3/internal/config"
	"github.com/stanislavpetrov/Go_Audi_MMI_mp3/internal/downloader"
)

var rootCmd = &cobra.Command{
	Use:   "ytmp3",
	Short: "YouTube to MP3 downloader with configurable audio parameters",
	Long: `A command-line tool to download YouTube videos as MP3 files
with configurable audio parameters such as bitrate, channels, and sample rate.`,
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download YouTube videos as MP3 files",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get configuration
		cfg, err := config.New()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get URLs from flag
		urlsStr, _ := cmd.Flags().GetString("urls")
		if urlsStr == "" {
			return fmt.Errorf("no URLs provided")
		}

		// Parse URLs
		urls, err := config.ParseURLs(urlsStr)
		if err != nil {
			return fmt.Errorf("failed to parse URLs: %w", err)
		}
		cfg.URLs = urls

		// Create context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\nReceived interrupt signal. Shutting down gracefully...")
			cancel()
		}()

		// Create downloader and start batch download
		dl := downloader.New(cfg)
		return dl.DownloadBatch(ctx, cfg.URLs)
	},
}

func init() {
	// Add download command flags
	downloadCmd.Flags().String("urls", "", "Comma-separated list of YouTube URLs to download")
	downloadCmd.Flags().Int("bitrate", 320, "Audio bitrate in kbps (1-320)")
	downloadCmd.Flags().Int("channels", 2, "Number of audio channels (1 or 2)")
	downloadCmd.Flags().Int("sample-rate", 48000, "Audio sample rate in Hz")
	downloadCmd.Flags().String("output-dir", "./downloads", "Output directory for downloaded files")

	// Bind flags to environment variables
	downloadCmd.Flags().Set("bitrate", os.Getenv("YTMP3_BITRATE"))
	downloadCmd.Flags().Set("channels", os.Getenv("YTMP3_CHANNELS"))
	downloadCmd.Flags().Set("sample-rate", os.Getenv("YTMP3_SAMPLE_RATE"))
	downloadCmd.Flags().Set("output-dir", os.Getenv("YTMP3_OUTPUT_DIR"))

	rootCmd.AddCommand(downloadCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
