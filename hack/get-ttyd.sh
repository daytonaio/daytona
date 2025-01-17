#!/bin/bash
# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

RELEASE_TAG="1.7.7"
RELEASE_ORG="tsl0922"
TTYD_ROOT="$HOME/ttyd"

# Check if ttyd is already installed
if [ -d "$TTYD_ROOT" ]; then
  echo "Terminal Server is already installed. Skipping installation."
  exit 0
fi

# Ensure the RELEASE_TAG is set
if [ -z "$RELEASE_TAG" ]; then
  echo "The RELEASE_TAG build arg must be set." >&2
  exit 1
fi

# Determine system architecture
arch=$(uname -m)
if [ "$arch" = "x86_64" ]; then
  arch="x86_64"
elif [ "$arch" = "aarch64" ]; then
  arch="aarch64"
elif [ "$arch" = "armv7l" ]; then
  arch="armhf"
else
  echo "Unsupported architecture: $arch"
  exit 1
fi

# Define the download URL and target file
download_url="https://github.com/$RELEASE_ORG/ttyd/releases/download/$RELEASE_TAG/ttyd.$arch"
target_file="$HOME/ttyd-$arch"

# Download the file using wget or curl
if command -v wget &>/dev/null; then
  wget -O "$target_file" "$download_url"
elif command -v curl &>/dev/null; then
  curl -fsSL -o "$target_file" "$download_url"
else
  echo "Neither wget nor curl is available. Please install one of them."
  exit 1
fi

# Make the binary executable
chmod +x "$target_file"

# Move ttyd to the installation directory
mkdir -p "$TTYD_ROOT/bin"
mv "$target_file" "$TTYD_ROOT/bin/ttyd"