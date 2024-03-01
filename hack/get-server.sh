#!/bin/bash

VERSION="latest"
if [ -n "$DAYTONA_SERVER_VERSION" ]; then
  VERSION=$DAYTONA_SERVER_VERSION
fi

# Check machine architecture
ARCH=$(uname -m)
# Check operating system
OS=$(uname -s)

case $OS in
  "Darwin")
    FILENAME="darwin"
    ;;
  "Linux")
    FILENAME="linux"
    echo "Unsupported operating system: $OS"
    exit 1
    ;;
  *)
    echo "Unsupported operating system: $OS"
    exit 1
    ;;
esac

case $ARCH in
  "arm64" | "ARM64")
    FILENAME="$FILENAME-arm64"
    ;;
  "x86_64" | "AMD64")
    FILENAME="$FILENAME-amd64"
    ;;
  "aarch64")
    FILENAME="$FILENAME-arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

BASE_URL="https://download.daytona.io/daytona"
if [ -n "$DAYTONA_SERVER_DOWNLOAD_URL" ]; then
  BASE_URL=$DAYTONA_SERVER_DOWNLOAD_URL
fi

DOWNLOAD_URL="$BASE_URL/$VERSION/daytona-$FILENAME"

echo "Downloading server from $DOWNLOAD_URL"

sudo curl $DOWNLOAD_URL -Lo daytona
sudo mv daytona /usr/local/bin/daytona
sudo chmod +x /usr/local/bin/daytona
