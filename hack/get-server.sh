#!/bin/bash

# This script downloads the Daytona server binary and installs it to /usr/local/bin
# You can set the environment variable DAYTONA_SERVER_VERSION to specify the version to download
# You can set the environment variable DAYTONA_SERVER_DOWNLOAD_URL to specify the base URL to download from
# You can set the environment variable DAYTONA_PATH to specify the path where to install the binary

VERSION=${DAYTONA_SERVER_VERSION:-"latest"}
BASE_URL=${DAYTONA_SERVER_DOWNLOAD_URL:-"https://download.daytona.io/daytona"}
DESTINATION=${DAYTONA_PATH:-"/usr/local/bin"}

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

if [ ! "$DAYTONA_PATH" ]; then
  echo "Default installation directory: /usr/local/bin"
  echo "You can override this by setting the DAYTONA_PATH environment variable (ie. \`| DAYTONA_PATH=/home/user/bin bash\`)"
fi

# Check if destination exists and is writable
if [[ ! -d $DESTINATION ]]; then
  # Inform user about missing directory or write permissions
  echo -e "\nWarning: Destination directory $DESTINATION does not exist."
  # Provide instructions on how to create dir
  echo "         Create the directory:"
  echo "           mkdir -p $DESTINATION"
  exit 1
fi
if [[ ! -w $DESTINATION ]]; then
  echo -e "\nWarning: Destination directory $DESTINATION is not writeable."
  echo "         Rerun the script with SUDO privileges:"
  if [ "$DAYTONA_PATH" ]; then
    echo "           curl -sf -L https://download.daytona.io/daytona/get-server.sh | DAYTONA_PATH=$DESTINATION sudo bash"
  else
    echo "           curl -sf -L https://download.daytona.io/daytona/get-server.sh | sudo bash"
  fi
  exit 1
fi

DOWNLOAD_URL="$BASE_URL/$VERSION/daytona-$FILENAME"

echo -e "\nDownloading server from $DOWNLOAD_URL"

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
if ! curl -fsSL "$DOWNLOAD_URL" -o "$temp_file"; then
  err "Daytona server download failed"
fi
chmod +x "$temp_file"

echo "Installing server to $DESTINATION"
mv "$temp_file" "$DESTINATION/daytona"

# Check if destination is in user's PATH
if [[ ! :"$PATH:" == *":$DESTINATION:"* ]]; then
  echo -e "\nWarning: $DESTINATION is not currently in your PATH environment variable."
  echo "         To be able to run the Daytona server from any directory, you may need to add it to your PATH."
  echo "         Edit your shell configuration file (e.g., ~/.bashrc or ~/.zshrc)"
  echo "         Add the following line:"
  echo "             export PATH=\$PATH:$DESTINATION"
  echo "         Source the configuration file to apply the changes (e.g., source ~/.bashrc)"
  else
  echo -e "\nRunning the Daytona server in daemon mode"
  daytona server -d
fi
