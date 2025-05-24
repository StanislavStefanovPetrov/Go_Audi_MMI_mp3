.PHONY: build install clean test

# Build the application
build:
	go build -o bin/ytmp3 cmd/main.go

# Install the application
install:
	go install ./cmd/main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Check if ffmpeg is installed
check-ffmpeg:
	@which ffmpeg >/dev/null 2>&1 || (echo "Error: ffmpeg is not installed. Please install ffmpeg first." && exit 1)

# Default target
all: check-ffmpeg build 