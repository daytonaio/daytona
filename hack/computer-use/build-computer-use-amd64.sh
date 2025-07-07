#!/bin/bash

set -e

# Check if current architecture is amd64
if [ "$(uname -m)" = "x86_64" ]; then
    echo "Building computer-use for amd64 architecture..."
    cd libs/computer-use
    go build -o ../../dist/libs/computer-use-amd64 main.go
    echo "Build completed successfully"
    exit 0
fi

# Build using docker image builder
docker build --platform linux/amd64 -t computer-use-amd64:build -f hack/computer-use/Dockerfile --no-cache .

# Run the container to copy the amd binary
docker run --rm -v $(pwd)/dist:/dist computer-use-amd64:build