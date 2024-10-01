#!/bin/bash

# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# This script downloads the Daytona binary and installs it to /usr/local/bin
# You can set the environment variable DAYTONA_SERVER_VERSION to specify the version to download
# You can set the environment variable DAYTONA_SERVER_DOWNLOAD_URL to specify the base URL to download from
# You can set the environment variable DAYTONA_PATH to specify the path where to install the binary

VERSION=${DAYTONA_SERVER_VERSION:-"latest"}
BASE_URL=${DAYTONA_SERVER_DOWNLOAD_URL:-"https://download.daytona.io/daytona"}
DESTINATION=${DAYTONA_PATH:-"/usr/local/bin"}
CONFIRM_FLAG=false

# Check for the -y flag
for arg in "$@"; do
  case $arg in
  -y)
    CONFIRM_FLAG=true
    shift
    ;;
  esac
done

# Print error message to stderr and exit
err() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" >&2
  exit 1
}

# Get the user that is running the Daytona server
get_daytona_server_user() {
    local user
    user=$(ps aux | grep 'daytona serve' | grep -v 'grep' | awk '{print $1}' | head -n 1)

    if [ -z "$user" ]; then
        echo "Error: daytona serve process not found" >&2
        return 1
    fi

    echo "$user"
}

# Stop the Daytona server if it is running
stop_daytona_server() {
  local pid
  pid=$(pgrep -f "daytona serve" | head -n 1)

  if [ ! -z "$pid" ]; then
    if [ "$CONFIRM_FLAG" = false ]; then
      read -p "Daytona server is running. Do you want to stop it? (yes/no): " user_input </dev/tty
      case $user_input in
      [Yy][Ee][Ss])
        CONFIRM_FLAG=true
        ;;
      [Nn][Oo])
        echo "Please stop the Daytona server manually and rerun the script."
        exit 0
        ;;
      *)
        echo "Invalid input. Please enter 'yes' or 'no'."
        exit 1
        ;;
      esac
    fi

    if [ "$CONFIRM_FLAG" = true ]; then
      echo "Attempting to stop the Daytona server..."
      user=$(get_daytona_server_user)
      if sudo -H -E -u $user bash -c 'daytona server stop'; then
        echo "Daytona server stopped successfully."
      else
        echo "Failed to stop Daytona server gracefully. Attempting force stop on process $pid..."
        if kill -9 $pid; then
          echo "Daytona server forcefully stopped."
        else
          echo "Failed to stop Daytona server. Please stop it manually and rerun the script."
          exit 1
        fi
      fi
    fi
  fi
}

# Check if the Daytona server is running
stop_daytona_server

echo -e "Installing Daytona...\n"

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
    echo "           curl -sf -L https://download.daytona.io/daytona/install.sh | DAYTONA_PATH=$DESTINATION sudo bash"
  else
    echo "           curl -sf -L https://download.daytona.io/daytona/install.sh | sudo bash"
  fi
  exit 1
fi

DOWNLOAD_URL="$BASE_URL/$VERSION/daytona-$FILENAME"

echo -e "\nDownloading Daytona binary from $DOWNLOAD_URL"

# Create a temporary file to download the Daytona binary. Just in case the user
# has file named "daytona" in the current directory.
temp_file="daytona-$RANDOM"

# Ensure the temporary file is deleted on exit
trap 'rm -f "$temp_file"' EXIT

# Function to determine file size in a cross-platform way
get_file_size() {
  local file=$1
  case "$(uname)" in
  Darwin)
    stat -f "%z" "$file" 2>/dev/null || echo "0"
    ;;
  Linux)
    stat --printf="%s" "$file" 2>/dev/null || echo "0"
    ;;
  *)
    echo "0"
    ;;
  esac
}

# Function to display a progress bar with a spinner
progress_bar_with_spinner() {
  local current_size=$1
  local total_size=$2
  local percent=$((current_size * 100 / total_size))
  local bar_width=40
  local filled=$(((percent * bar_width + 99) / 100))
  local empty=$((bar_width - filled))
  local spinner="⠋⠙⠸⠴⠦⠧⠇⠏"
  local spin_char=${spinner:$(($RANDOM % ${#spinner})):1}

  # Generate the filled portion of the bar
  local bar=$(printf "%0.s#" $(seq 1 $filled))

  # Only add the empty portion if the download is not complete
  if [ $percent -lt 100 ]; then
    bar+=$(printf "%0.s-" $(seq 1 $empty))
  fi
  printf "\r\033[K[%s] Downloading... [%-${bar_width}s] %d%%" "$spin_char" "$bar" "$percent"
}

# Download the file while showing a progress bar with a spinner
download_file() {
  local url=$1
  local output=$2

  # Get the total size of the file to download
  local total_size=$(curl -sI "$url" | grep -i Content-Length | awk '{print $2}' | tr -d '\r')

  # Initialize a temporary file to hold the download
  local temp_output=$(mktemp)

  # Start downloading the file with `curl` and pipe the output to `dd` to track progress
  #
  # Flags:
  # -f, --fail: Fail silently on server errors.
  # -s, --silent: Silent mode. Don't show progress meter or error messages.
  # -S, --show-error: When used with -s, show error even if silent mode is enabled.
  # -L, --location: Follow redirects.
  # -o, --output <file>: Write output to <file> instead of stdout.
  curl -fsSL "$url" -o "$temp_output" &
  local curl_pid=$!

  while kill -0 $curl_pid 2>/dev/null; do
    local downloaded_size=$(get_file_size "$temp_output")
    if [ "$downloaded_size" -gt 0 ]; then
      progress_bar_with_spinner "$downloaded_size" "$total_size"
    fi
    sleep 0.1
  done

  # Ensure the progress bar reaches 100% on successful completion
  progress_bar_with_spinner "$total_size" "$total_size"
  printf "\n"

  mv "$temp_output" "$output"
  echo ""
}

# curl does not fail with a non-zero status code when the HTTP status
# code is not successful (like 404 or 500). You can add -f flag to curl
# command to fail silently on server errors.

if ! download_file "$DOWNLOAD_URL" "$temp_file"; then
  err "Daytona binary download failed"
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
fi

echo -e "\nDaytona has been successfully installed to $DESTINATION!"
