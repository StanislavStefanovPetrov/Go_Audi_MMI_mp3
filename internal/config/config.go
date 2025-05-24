package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env"
)

// Config holds all configuration parameters for the application
type Config struct {
	// Audio settings
	Bitrate    int    `env:"YTMP3_BITRATE" envDefault:"320"`
	Channels   int    `env:"YTMP3_CHANNELS" envDefault:"2"`
	SampleRate int    `env:"YTMP3_SAMPLE_RATE" envDefault:"48000"`
	OutputDir  string `env:"YTMP3_OUTPUT_DIR" envDefault:"./downloads"`

	// Application settings
	MaxConcurrentDownloads int    `env:"YTMP3_MAX_CONCURRENT_DOWNLOADS" envDefault:"3"`
	TempDir                string `env:"YTMP3_TEMP_DIR" envDefault:""`
	LogLevel               string `env:"YTMP3_LOG_LEVEL" envDefault:"info"`

	// URLs to download (not from env, set via command line)
	URLs []string
}

// New creates a new Config instance with values from environment variables
func New() (*Config, error) {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate config
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Set default temp directory if not specified
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}

	return config, nil
}

func validateConfig(cfg *Config) error {
	// Validate bitrate
	if cfg.Bitrate <= 0 || cfg.Bitrate > 320 {
		return fmt.Errorf("invalid bitrate: must be between 1 and 320 kbps")
	}

	// Validate channels
	if cfg.Channels != 1 && cfg.Channels != 2 {
		return fmt.Errorf("invalid channels: must be either 1 (mono) or 2 (stereo)")
	}

	// Validate sample rate
	if cfg.SampleRate <= 0 {
		return fmt.Errorf("invalid sample rate: must be greater than 0")
	}

	// Validate output directory
	if cfg.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for output directory: %w", err)
	}
	cfg.OutputDir = absPath

	// Validate max concurrent downloads
	if cfg.MaxConcurrentDownloads <= 0 {
		return fmt.Errorf("max concurrent downloads must be greater than 0")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[strings.ToLower(cfg.LogLevel)] {
		return fmt.Errorf("invalid log level: must be one of debug, info, warn, error")
	}

	return nil
}

// ParseURLs splits a comma-separated string of URLs into a slice
func ParseURLs(urlsStr string) []string {
	if urlsStr == "" {
		return nil
	}
	return strings.Split(urlsStr, ",")
}
