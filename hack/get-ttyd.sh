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

# Download and set up ttyd
wget https://github.com/$RELEASE_ORG/ttyd/releases/download/$RELEASE_TAG/ttyd.$arch -O $HOME/ttyd-$arch
chmod +x $HOME/ttyd-$arch

# Move ttyd to installation directory
mkdir -p $TTYD_ROOT/bin
mv $HOME/ttyd-$arch $TTYD_ROOT/bin/ttyd

