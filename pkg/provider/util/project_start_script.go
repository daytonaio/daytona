// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "fmt"

func GetProjectStartScript(daytonaDownloadUrl string, apiKey string) string {
	return fmt.Sprintf(`
  # Check for missing dependencies
  missing_deps=""
  command -v sudo > /dev/null 2>&1 || missing_deps="$missing_deps sudo"
  command -v curl > /dev/null 2>&1 || missing_deps="$missing_deps curl"
  command -v bash > /dev/null 2>&1 || missing_deps="$missing_deps bash"
  command -v git > /dev/null 2>&1 || missing_deps="$missing_deps git"
  
  # Print missing dependencies
  if [ -n "$missing_deps" ]; then
    echo "Missing dependencies: $missing_deps"
  fi
  
  # Install missing dependencies if any
  if [ -n "$missing_deps" ]; then
    if command -v apt-get > /dev/null 2>&1; then
      echo "Installing missing dependencies using apt-get..."
      apt-get update && apt-get install -y $missing_deps > /dev/null 2>&1
    elif command -v yum > /dev/null 2>&1; then
      echo "Installing missing dependencies using yum..."
      yum install -y $missing_deps > /dev/null 2>&1
    elif command -v apk > /dev/null 2>&1; then
      echo "Installing missing dependencies using apk..."
      apk add --no-cache $missing_deps libc6-compat > /dev/null 2>&1
    elif command -v dnf > /dev/null 2>&1; then
      echo "Installing missing dependencies using dnf..."
      dnf install -y $missing_deps > /dev/null 2>&1
    elif command -v brew > /dev/null 2>&1; then
      echo "Installing missing dependencies using brew..."
      brew install $missing_deps > /dev/null 2>&1
    elif command -v pacman > /dev/null 2>&1; then
      echo "Installing missing dependencies using pacman..."
      pacman -Sy --noconfirm $missing_deps > /dev/null 2>&1
    else
      echo "Cannot install missing dependencies: $missing_deps"
      exit 1
    fi
  fi

  # Download and install Daytona agent
  curl -sfL -H "Authorization: Bearer %s" %s | sudo -E bash && daytona agent
  `, apiKey, daytonaDownloadUrl)
}
