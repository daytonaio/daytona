#!/bin/bash

VERSION="latest"
if [ -n "$DAYTONA_SERVER_VERSION" ]; then
  VERSION=$DAYTONA_SERVER_VERSION
fi 
  
# Check machine architecture
ARCH=$(uname -m)
# Check operating system
OS=$(uname -s)

if [ "$OS" == "" ]; then
  OS=$(ver)
  ARCH=$(echo %PROCESSOR_ARCHITECTURE%)
fi

if [ "$OS" == "Darwin" ]; then
  FILENAME="darwin"
elif [ "$OS" == "Linux" ]; then
  FILENAME="linux"
elif [[ $OS == *"Windows"* ]]; then
  FILENAME="windows"
else
  echo "Unsupported operating system: $OS"
  exit 1
fi

if [ "$ARCH" == "arm64" ]; then
  FILENAME=$(echo "$FILENAME-arm64")
elif [ "$ARCH" == "ARM64" ]; then
  FILENAME=$(echo "$FILENAME-arm64")
elif [ "$ARCH" == "x86_64" ]; then
  FILENAME=$(echo "$FILENAME-amd64")
elif [ "$ARCH" == "AMD64" ]; then
  FILENAME=$(echo "$FILENAME-amd64")
elif [ "$ARCH" == "aarch64" ]; then
  FILENAME=$(echo "$FILENAME-arm64")
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

BASE_URL="https://download.daytona.io/daytona"
if [ -n "$DAYTONA_SERVER_DOWNLOAD_URL" ]; then
  BASE_URL=$DAYTONA_SERVER_DOWNLOAD_URL
fi

DOWNLOAD_URL=$(echo "$BASE_URL/$VERSION/daytona-$FILENAME")

echo "Downloading server from $DOWNLOAD_URL"

sudo curl $DOWNLOAD_URL -Lo daytona
sudo mv daytona /usr/local/bin/daytona
sudo chmod +x /usr/local/bin/daytona