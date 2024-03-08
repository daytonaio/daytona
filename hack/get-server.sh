#!/bin/bash

# This script downloads the Daytona server binary and installs it to /usr/local/bin
# You can set the environment variable DAYTONA_SERVER_VERSION to specify the version to download
# You can set the environment variable DAYTONA_SERVER_DOWNLOAD_URL to specify the base URL to download from

VERSION=${DAYTONA_SERVER_VERSION:-"latest"}
BASE_URL=${DAYTONA_SERVER_DOWNLOAD_URL:-"https://download.daytona.io/daytona"}
DESTINATION=""
SUDO_REQUIRED=0

# Print error message to stderr and exit
err() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" >&2
  exit 1
}

# Check if there is a directory in $PATH that we can write to. With this
# statement the first match from top to bottom will be used.
# shellcheck disable=SC2088
case :$PATH: in
  *:$HOME/bin:*)
    DESTINATION="$HOME/bin"
    SUDO_REQUIRED=0
    ;;
  *:$HOME/.local/bin:*)
    DESTINATION="$HOME/.local/bin"
    SUDO_REQUIRED=0
    ;;
  *:/usr/local/bin:*)
    DESTINATION="/usr/local/bin"
    SUDO_REQUIRED=1
    ;;
  *:/opt/bin:*)
    DESTINATION="/opt/bin"
    SUDO_REQUIRED=1
    ;;
  *) err "~/bin, ~/.local/bin, /opt/bin and /usr/local/bin not on PATH. No option to install to.";;
esac

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

# We want to fail fast so check everything before we download

# Check if the $DESTINATION exists. We are making a decision to not mkdir
# for the user here. /opt/bin and /usr/local/bin are system directories and
# the user should decide and explicitly create those.
if [ ! -d "$DESTINATION" ]; then
  err "Destination directory $DESTINATION does not exist. Run mkdir -p $DESTINATION and re-run."
fi

# Check if sudo is required
if [ "$SUDO_REQUIRED" -eq 1 ] && [ "$EUID" -ne 0 ]; then
  err "Cannot write to /opt/bin or /usr/local/bin as non root. Please re-run with sudo."
fi

echo "Downloading server from $DOWNLOAD_URL"

# Create a temporary file to download the server binary. Just in case the author... user
# has say a directory named "daytona" in $HOME... the current directory.
temp_file="daytona-$RANDOM"

# Ensure the temporary file is deleted on exit
trap 'rm -f "$temp_file"' EXIT

# curl does not fail with a non-zero status code when the HTTP status
# code is not successful (like 404 or 500). You can add -f flag to curl
# command to fail silently on server errors.
#
# Flags:
# -f, --fail: Fail silently on server errors.
# -s, --silent: Silent mode. Don't show progress meter or error messages.
# -S, --show-error: When used with -s, show error even if silent mode is enabled.
# -L, --location: Follow redirects.
# -o, --output <file>: Write output to <file> instead of stdout.
curl -fsSL "$DOWNLOAD_URL" -o "$temp_file"
if [ $? -ne 0 ]; then
  err "Daytona server download failed"
fi
chmod +x "$temp_file"

echo "Installing server to $DESTINATION"
mv "$temp_file" "$DESTINATION/daytona"
