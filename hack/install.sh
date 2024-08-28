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

# Check if the terminal supports ANSI escape codes
if [[ $TERM == *256color* || $TERM == *xterm* ]]; then
  RESET='\033[0m'
  BOLD='\033[1m'
  BLUE_BG='\033[44m'  # Blue background
  BLUE_FG='\033[34m'  # Blue foreground
else
  RESET=''
  BOLD=''
  BLUE_BG=''  # No color support
  BLUE_FG=''
fi

# Function to draw the animated progress bar with a blue background and bold text
draw_animated_bar() {
  local progress=$1
  local total=$2
  local speed=$3
  local width=50
  local completed=$((progress * width / total))
  local remaining=$((width - completed))

  # Limit progress to avoid overshooting 100%
  if [ $progress -gt $total ]; then
    progress=$total
  fi

  # Calculate percentage
  local percentage=$((progress * 100 / total))
  [ $percentage -gt 100 ] && percentage=100

  # Draw progress bar
  printf "\r\033[K["
  for ((i=0; i<completed; i++)); do
    printf "${BLUE_FG}░${RESET}"
  done
  for ((i=0; i<remaining; i++)); do
    printf "░"
  done
  printf "] ${BOLD}${percentage}%% ${speed}${RESET}"
}

# Stop the Daytona server if it is running
stop_daytona_server() {
  if pgrep -x "daytona" > /dev/null; then
    if [ "$CONFIRM_FLAG" = false ]; then
      read -p "Daytona server is running. Do you want to stop it? (yes/no): " user_input < /dev/tty
      case $user_input in
        [Yy][Ee][Ss] )
          CONFIRM_FLAG=true
          ;;
        [Nn][Oo] )
          echo "Please stop the Daytona server manually and rerun the script."
          exit 0
          ;;
        * )
          echo "Invalid input. Please enter 'yes' or 'no'."
          exit 1
          ;;
      esac
    fi

    if [ "$CONFIRM_FLAG" = true ]; then
      echo "Attempting to stop the Daytona server..."
      if daytona server stop; then
        echo -e "Stopping the daytona server"
      else
        pkill -x "daytona"
      fi
      echo -e "Daytona server stopped.\n"
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

# Function to download the file with progress bar and speed
download_file() {
  local url=$1
  local output=$2

  # Get the total size of the file to download
  local total_size=$(curl -sI "$url" | grep -i Content-Length | awk '{print $2}' | tr -d '\r')

  # Initialize download start time
  local start_time=$(date +%s)

  # Download the file with curl
  curl -fsSL "$url" -o "$temp_file" &
  local curl_pid=$!

  while kill -0 $curl_pid 2>/dev/null; do
    local downloaded_size=$(stat -c %s "$temp_file")
    local current_time=$(date +%s)
    local elapsed_time=$((current_time - start_time))

    # Ensure elapsed_time is not zero to avoid division by zero errors
    if [ $elapsed_time -eq 0 ]; then
      elapsed_time=1
    fi

    # Calculate download speed (in bytes per second)
    local speed=$(echo "scale=2; $downloaded_size / $elapsed_time" | bc)
    local formatted_speed
    if [ "$(echo "$speed >= 1048576" | bc)" -eq 1 ]; then
      formatted_speed=$(echo "scale=2; $speed / 1048576" | bc)MB/s
    elif [ "$(echo "$speed >= 1024" | bc)" -eq 1 ]; then
      formatted_speed=$(echo "scale=2; $speed / 1024" | bc)KB/s
    else
      formatted_speed=$(echo "scale=2; $speed" | bc)B/s
    fi

    draw_animated_bar "$downloaded_size" "$total_size" "$formatted_speed"
    sleep 0.1
  done

  mv "$temp_file" "$output"
  echo ""
}

# Download the Daytona binary
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
