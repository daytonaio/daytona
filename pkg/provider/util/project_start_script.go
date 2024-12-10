// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "fmt"

func GetProjectStartScript(daytonaDownloadUrl string, apiKey string) string {
	return fmt.Sprintf(`
# List of supported package managers
PACKAGE_MANAGERS="apt-get yum dnf apk brew pacman"

# Ensure sudo is installed
if ! command -v sudo > /dev/null 2>&1; then
  for pm in $PACKAGE_MANAGERS; do
    if command -v "$pm" > /dev/null 2>&1; then
      case "$pm" in
        apt-get)
          apt-get update
          apt-get install -y sudo > /dev/null 2>&1
          ;;
        yum)
          yum install -y sudo > /dev/null 2>&1
          ;;
        dnf)
          dnf install -y sudo > /dev/null 2>&1
          ;;
        apk)
          apk add --no-cache sudo > /dev/null 2>&1
          ;;
        brew)
          brew install sudo > /dev/null 2>&1
          ;;
        pacman)
          pacman -Sy --noconfirm sudo > /dev/null 2>&1
          ;;
      esac
      
      # Break after first successful installation
      [ $? -eq 0 ] && break
    fi
  done
fi

# Verify sudo is working
if ! sudo -v; then
  echo "Failed to configure sudo. Check system permissions."
  exit 1
fi

# Check for missing dependencies
DEPENDENCIES="curl bash git"
MISSING_DEPS=""

for dep in $DEPENDENCIES; do
  if ! command -v "$dep" > /dev/null 2>&1; then
    MISSING_DEPS="$MISSING_DEPS $dep"
  fi
done

# Install missing dependencies
if [ -n "$MISSING_DEPS" ]; then
  echo "Missing dependencies:$MISSING_DEPS"
  
  for pm in $PACKAGE_MANAGERS; do
    if command -v "$pm" > /dev/null 2>&1; then
      case "$pm" in
        apt-get)
          sudo apt-get update
          sudo apt-get install -y $MISSING_DEPS > /dev/null 2>&1
          ;;
        yum)
          sudo yum install -y $MISSING_DEPS > /dev/null 2>&1
          ;;
        dnf)
          sudo dnf install -y $MISSING_DEPS > /dev/null 2>&1
          ;;
        apk)
          sudo apk add --no-cache $MISSING_DEPS > /dev/null 2>&1
          ;;
        brew)
          sudo brew install $MISSING_DEPS > /dev/null 2>&1
          ;;
        pacman)
          sudo pacman -Sy --noconfirm $MISSING_DEPS > /dev/null 2>&1
          ;;
      esac
      
      # Break after first successful installation
      [ $? -eq 0 ] && break
    fi
  done
fi

# Download and install Daytona agent
curl -sfL -H "Authorization: Bearer %s" %s | sudo -E sh && daytona agent
`, apiKey, daytonaDownloadUrl)
}
