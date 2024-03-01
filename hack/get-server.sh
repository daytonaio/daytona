#!/bin/bash

# This script downloads the Daytona server binary and installs it to /usr/local/bin
# You can set the environment variable DAYTONA_SERVER_VERSION to specify the version to download
# You can set the environment variable DAYTONA_SERVER_DOWNLOAD_URL to specify the base URL to download from

VERSION=${DAYTONA_SERVER_VERSION:-"latest"}
BASE_URL=${DAYTONA_SERVER_DOWNLOAD_URL:-"https://download.daytona.io/daytona"}
DESTINATION="/usr/local/bin/daytona"

# Print error message to stderr and exit
err() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" >&2
  exit 1
}

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
    ;;
  *)
    err "Unsupported operating system: $OS"
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
    err "Unsupported architecture: $ARCH"
    ;;
esac

DOWNLOAD_URL="$BASE_URL/$VERSION/daytona-$FILENAME"

echo "Downloading server from $DOWNLOAD_URL"

# curl does not fail with a non-zero status code when the HTTP status
# code is not successful (like 404 or 500). You can add -f flag to curl
# command to fail silently on server errors.
curl -fsSL "$DOWNLOAD_URL" -o daytona
if [ $? -ne 0 ]; then
  err "Error occurred while downloading the server"
fi

sudo mv daytona "$DESTINATION"
sudo chmod +x "$DESTINATION"
