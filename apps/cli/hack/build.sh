#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0


# Exit on error
set -e

# Get absolute path of script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST_DIR="$(cd "${SCRIPT_DIR}/../../.." && pwd)"

# Environment file precedence:
# 1. DAYTONA_ENV_FILE environment variable if set
# 2. .env file in CLI directory
# 3. .env file in project root
# 4. Default values

load_env_file() {
    local env_file="$1"
    if [ -f "$env_file" ]; then
        source "$env_file"
        return 0
    fi
    return 1
}

# Try loading environment files in order of precedence
if [ -n "$DAYTONA_ENV_FILE" ]; then
    if ! load_env_file "$DAYTONA_ENV_FILE"; then
        echo "Warning: Environment file specified by DAYTONA_ENV_FILE ($DAYTONA_ENV_FILE) not found"
    fi
elif load_env_file "${SCRIPT_DIR}/../.env.local"; then
    : # Successfully loaded CLI .env
elif load_env_file "${SCRIPT_DIR}/../.env"; then
    : # Successfully loaded CLI .env
elif load_env_file "${PROJECT_ROOT}/.env.local"; then
    : # Successfully loaded root .env
elif load_env_file "${PROJECT_ROOT}/.env"; then
    : # Successfully loaded root .env
else
    echo "Note: No .env file found, using default values"
fi


# Set default values
: "${DAYTONA_VERSION:=v0.0.0-dev}"
: "${GOOS:=linux}"
: "${GOARCH:=amd64}"
: "${CGO_ENABLED:=0}"

# Export for build
export DAYTONA_VERSION
export GOOS
export GOARCH
export CGO_ENABLED

# Validate required variables
REQUIRED_VARS=(
    "DAYTONA_API_URL"
    "DAYTONA_AUTH0_DOMAIN"
    "DAYTONA_AUTH0_CLIENT_ID"
    "DAYTONA_AUTH0_CALLBACK_PORT"
    "DAYTONA_AUTH0_AUDIENCE"
)

MISSING_VARS=()
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        MISSING_VARS+=("$var")
    fi
done

if [ ${#MISSING_VARS[@]} -ne 0 ]; then
    echo "Error: Missing required environment variables:"
    printf '%s\n' "${MISSING_VARS[@]}"
    exit 1
fi

# Create build directory if it doesn't exist
mkdir -p "${DIST_DIR}/dist/apps/cli"

# Build the binary
echo "Building Daytona CLI with version: $DAYTONA_VERSION"
go build \
    -ldflags "-X 'github.com/daytonaio/daytona/cli/internal.Version=${DAYTONA_VERSION}' \
    -X 'github.com/daytonaio/daytona/cli/internal.DaytonaApiUrl=${DAYTONA_API_URL}' \
    -X 'github.com/daytonaio/daytona/cli/internal.Auth0Domain=${DAYTONA_AUTH0_DOMAIN}' \
    -X 'github.com/daytonaio/daytona/cli/internal.Auth0ClientId=${DAYTONA_AUTH0_CLIENT_ID}' \
    -X 'github.com/daytonaio/daytona/cli/internal.Auth0ClientSecret=${DAYTONA_AUTH0_CLIENT_SECRET}' \
    -X 'github.com/daytonaio/daytona/cli/internal.Auth0CallbackPort=${DAYTONA_AUTH0_CALLBACK_PORT}' \
    -X 'github.com/daytonaio/daytona/cli/internal.Auth0Audience=${DAYTONA_AUTH0_AUDIENCE}'" \
    -o "${DIST_DIR}/dist/apps/cli/daytona-${GOOS}-${GOARCH}" main.go

echo "Build complete: ${DIST_DIR}/dist/apps/cli/daytona-${GOOS}-${GOARCH}"
