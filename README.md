# YouTube to MP3 Downloader

A Go application that downloads YouTube videos as MP3 files with configurable audio parameters.

## Features

- Download YouTube videos as MP3 files
- Configurable audio parameters:
  - Bitrate (default: 320 kbps)
  - Channels (default: 2)
  - Sample rate (default: 48000 Hz)
- Support for batch downloading from a list of URLs
- Configurable target folder for downloads

## Installation

```bash
go install github.com/stanislavpetrov/Go_Audi_MMI_mp3@latest
```

## Usage

### Command Line Arguments

```bash
# Basic usage with default parameters
ytmp3 download --urls "https://www.youtube.com/watch?v=VIDEO_ID1,https://www.youtube.com/watch?v=VIDEO_ID2"

# With custom parameters
ytmp3 download --urls "https://www.youtube.com/watch?v=VIDEO_ID" \
    --bitrate 256 \
    --channels 2 \
    --sample-rate 44100 \
    --output-dir "./downloads"
```

### Environment Variables

You can also configure the application using environment variables:

```bash
export YTMP3_BITRATE=256
export YTMP3_CHANNELS=2
export YTMP3_SAMPLE_RATE=44100
export YTMP3_OUTPUT_DIR="./downloads"
```

## Configuration

The following parameters can be configured:

- `--bitrate`: Audio bitrate in kbps (default: 320)
- `--channels`: Number of audio channels (default: 2)
- `--sample-rate`: Audio sample rate in Hz (default: 48000)
- `--output-dir`: Target directory for downloads (default: "./downloads")
- `--urls`: Comma-separated list of YouTube URLs to download

## License

MIT License
