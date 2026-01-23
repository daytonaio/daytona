#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

set -e

echo "Installing Python packages and Go tools..."

USERNAME="${USERNAME:-"${_REMOTE_USER:-"automatic"}"}"

# Determine the appropriate non-root user
if [ "${USERNAME}" = "auto" ] || [ "${USERNAME}" = "automatic" ]; then
    USERNAME=""
    POSSIBLE_USERS=("vscode" "node" "codespace" "$(awk -v val=1000 -F ":" '$3==val{print $1}' /etc/passwd)")
    for CURRENT_USER in "${POSSIBLE_USERS[@]}"; do
        if id -u "${CURRENT_USER}" > /dev/null 2>&1; then
            USERNAME=${CURRENT_USER}
            break
        fi
    done
    if [ "${USERNAME}" = "" ]; then
        USERNAME=root
    fi
elif [ "${USERNAME}" = "none" ] || ! id -u "${USERNAME}" > /dev/null 2>&1; then
    USERNAME=root
fi

export GOROOT="${TARGET_GOROOT:-"/usr/local/go"}"
export GOPATH="${TARGET_GOPATH:-"/go"}"
export GOCACHE=/tmp/gotools/cache


sudo -E -u "${USERNAME}" bash -c '
export PATH=$GOROOT/bin:$PATH
export HOME=/home/${USER}

# Install pip packages
if [ -n "$PIPPACKAGES" ]; then
    echo "Installing pip packages: $PIPPACKAGES"
    IFS=',' read -ra PACKAGES <<< "${PIPPACKAGES}"
    pip3 install --no-cache-dir "${PACKAGES[@]}"
else
    echo "No pip packages specified. Skipping."
fi

# Install Go tools
if [ -n "$GOTOOLS" ]; then
    echo "Installing Go tools: $GOTOOLS"
    IFS=',' read -ra TOOLS <<< "${GOTOOLS}"
    for tool in "${TOOLS[@]}"; do
        go install $tool
    done
else
    echo "No Go tools specified. Skipping."
fi
'

# Set insecure registry
cat > /etc/docker/daemon.json <<EOF 
{
  "insecure-registries": ["registry:5000"]
}
EOF
