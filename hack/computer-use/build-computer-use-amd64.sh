#!/bin/bash

set -e

# Skip build if SKIP_COMPUTER_USE_BUILD is set
if [ -n "$SKIP_COMPUTER_USE_BUILD" ]; then
    echo "Skipping computer-use build"
    exit 0
fi

# Check if current architecture is amd64
if [ "$(uname -m)" = "x86_64" ]; then
    echo "Building computer-use for amd64 architecture (native build)..."
    cd libs/computer-use
    go build -o ../../dist/libs/computer-use-amd64 main.go
    echo "Native build completed successfully"
    exit 0
fi

echo "Current architecture: $(uname -m)"
echo "Building computer-use for amd64 architecture using Docker..."

# Ensure dist directory exists
mkdir -p dist/libs

# Build using docker image builder
echo "Building Docker image..."
docker build --platform linux/amd64 -t computer-use-amd64:build -f hack/computer-use/Dockerfile .

echo "Docker build completed, copying binary..."

# Run the container to copy the amd binary
docker run --rm -v "$(pwd)/dist:/dist" computer-use-amd64:build

# Verify the binary was created and show info
if [ -f "dist/libs/computer-use-amd64" ]; then
    echo "computer-use-amd64 build completed successfully"
    echo "Binary size: $(ls -lh dist/libs/computer-use-amd64 | awk '{print $5}')"
    echo "Binary location: $(pwd)/dist/libs/computer-use-amd64"
else
    echo "Error: Binary not found after build"
    exit 1
fi
