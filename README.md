# Audi YouTube Downloader (Go Edition)

A Go application that downloads audio files from YouTube in MP3 format, compatible with Audi Multi Media Interface (MMI) systems. You can listen to these tracks in your Audi car via SD/SDHC/MMC memory cards.


## About this project

This repo is created to assist Audi car owners with an MMI system in downloading and formatting music from YouTube so that the files are maximally compatible with the car's system. The reason is that Audi MMI imposes specific requirements on MP3 file formats, filenames, and metadata, which often leads to issues with song or album recognition. This tool automates the process and guarantees compatibility.

## Compatibility

This project takes into account the limitations of Audi MMI systems regarding track formats. It ensures that the downloaded audio files are in a format supported by Audi MMI systems. For more information on Audi MMI system limitations, refer to the following sources:

- [AudiWorld Forum: MMI 3G - Largest SD Card Size](https://www.audiworld.com/forums/q5-sq5-mki-8r-discussion-129/mmi-3g-largest-sd-card-size-2872958/#&gid=1&pid=1)


![Supported media and file format](https://github.com/StanislavStefanovPetrov/Audi_MMI_pytube_mp3/assets/29039888/371077bf-6104-48df-bf05-c8169dc06e2b)

## Tested on

- Audi MMI 3G+ (Audi Q7 4LB)
- Other models with MMI 3G/3G+ (compatibility is expected, but not personally tested).

## Prerequisites

- Go 1.21 or newer
- Installed yt-dlp and ffmpeg (see Makefile)

## Features

- Download audio files from YouTube videos as MP3 files and convert them into a format suitable for Audi MMI systems.
- Configurable audio parameters:
  - Bitrate (default: 320 kbps)
  - Channels (default: 2)
  - Sample rate (default: 48000 Hz)
- Support for a list of YouTube URLs for batch downloading.
- Organize downloaded files in a target folder for easy access.
- Sanitize metadata and filenames for maximum compatibility with Audi MMI.
- Removes (empties) the description and synopsis metadata tags from MP3 files to ensure maximum compatibility with Audi MMI systems.

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

## Notes

Feel free to customize and modify the application according to your needs.

## License

This project is licensed under the [MIT License](LICENSE).
