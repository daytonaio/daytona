// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"
)

const getDaytonaScript = `
#!/bin/bash

VERSION=${DAYTONA_SERVER_VERSION:-"latest"}
BASE_URL=${DAYTONA_SERVER_DOWNLOAD_URL:-"https://download.daytona.io/daytona"}
DESTINATION=${DAYTONA_PATH:-"/usr/local/bin"}

# Print error message to stderr and exit
err() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" >&2
  exit 1
}

try_download() {
  echo "Downloading Daytona from $1"

  local url=$1
  local temp_file=$2
  local exit_code=0

  if command -v wget > /dev/null 2>&1; then
    wget -q "$url" -O $temp_file --header="Authorization: Bearer $DAYTONA_SERVER_API_KEY" && return 0
    exit_code=$?
  elif command -v curl > /dev/null 2>&1; then
    curl -fsSL "$url" -H "Authorization: Bearer $DAYTONA_SERVER_API_KEY" -o "$temp_file" && return 0
    exit_code=$?
  else
    echo "error: Make sure curl or wget is available in the project container"
    exit 127
  fi
  >&2 echo "error: Daytona binary download failed. Exit Code: ${exit_code}"

  return 1
}

# Check if daytona is already installed
if [ -x "$(command -v daytona)" ]; then
  echo "Daytona already installed. Skipping installation..."
  exit 0
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

DIRECT_DOWNLOAD_URL=https://download.daytona.io/daytona/$VERSION/daytona-$FILENAME
DOWNLOAD_URL="$BASE_URL/$VERSION/daytona-$FILENAME"

if [[ ! "$DOWNLOAD_URL" =~ ^http://host.docker.internal ]]; then
  if curl -sfI "$DIRECT_DOWNLOAD_URL" > /dev/null; then
    DOWNLOAD_URL="$DIRECT_DOWNLOAD_URL"
  fi
fi

# Create a temporary file to download the Daytona binary. Just in case the user
# has file named "daytona" in the current directory.
temp_file="daytona-$RANDOM"

# Ensure the temporary file is deleted on exit
trap 'rm -f "$temp_file"' EXIT

i=1
max_retry=10
while :; do
  if try_download "$DOWNLOAD_URL" "$temp_file"; then
    break
  fi
  
  i=$((i+1))
  
  if [ "$i" -gt "$max_retry" ]; then
    >&2 echo "error: failed to download daytona after $max_retry attempts"
    exit 1
  fi

  >&2 echo "Trying again in 2 seconds..."
  sleep 2
done

echo "Daytona downloaded successfully"

chmod +x "$temp_file"

echo "Installing server to $DESTINATION"
mv "$temp_file" "$DESTINATION/daytona"

`

func GetDaytonaScript(baseUrl string) string {
	return strings.Replace(getDaytonaScript, "https://download.daytona.io/daytona", baseUrl, 1)
}
