#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Deploy runner-win to remote libvirt host and run in foreground
# This script builds locally, copies to the remote host, and executes
#
# Prerequisites:
# - SSH access to the libvirt host
# - SSH tunnels established via tunnel.sh (in a separate terminal)
#
# Usage:
#   DAYTONA_RUNNER_TOKEN=<token> ./deploy.sh
#
# Environment variables:
#   DAYTONA_RUNNER_TOKEN - Required: Authentication token for the runner
#   LIBVIRT_HOST        - Remote host (default: h1001.blinkbox.dev)
#   RUNNER_PORT         - Runner API port (default: 8080)
#   LOCAL_API_PORT      - Local API port for tunnel (default: 3000)
#   SKIP_BUILD          - Skip build step if set to "true"
#   DEPLOY_PATH         - Remote path to deploy binary (default: /tmp/daytona)

set -e

# Get script directory (works with both bash and when sourced)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Configuration with defaults
LIBVIRT_HOST="${LIBVIRT_HOST:-h1001.blinkbox.dev}"
RUNNER_PORT="${RUNNER_PORT:-8080}"
LOCAL_API_PORT="${LOCAL_API_PORT:-3001}"
DEPLOY_PATH="${DEPLOY_PATH:-/tmp/daytona}"
BINARY_NAME="runner-win"
BINARY_PATH="${WORKSPACE_ROOT}/dist/apps/${BINARY_NAME}"

echo "=== Daytona Runner-Win Deployment ==="
echo "Libvirt Host: $LIBVIRT_HOST"
echo "Runner Port: $RUNNER_PORT"
echo "Local API Port: $LOCAL_API_PORT"
echo "Deploy Path: $DEPLOY_PATH"
echo ""

# Check required environment variables
if [ -z "$DAYTONA_RUNNER_TOKEN" ]; then
    echo "Error: DAYTONA_RUNNER_TOKEN environment variable is required"
    echo ""
    echo "Usage: DAYTONA_RUNNER_TOKEN=<token> $0"
    exit 1
fi

# Build phase
if [ "$SKIP_BUILD" != "true" ]; then
    echo "=== Building runner-win ==="
    cd "$WORKSPACE_ROOT"
    npx nx build runner-win
    echo "✓ Build complete"
    echo ""
else
    echo "=== Skipping build (SKIP_BUILD=true) ==="
    echo ""
fi

# Verify binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH"
    echo "Run 'nx build runner-win' first or remove SKIP_BUILD=true"
    exit 1
fi

echo "=== Deploying to $LIBVIRT_HOST ==="

# Create deploy directory on remote
echo "Creating deploy directory..."
ssh "$LIBVIRT_HOST" "mkdir -p $DEPLOY_PATH"

# Stop any existing runner process
echo "Stopping existing runner process (if any)..."
ssh "$LIBVIRT_HOST" "pkill -f '$DEPLOY_PATH/$BINARY_NAME' 2>/dev/null || true"
sleep 1

# Copy binary to remote host
echo "Copying binary to remote host..."
scp "$BINARY_PATH" "$LIBVIRT_HOST:$DEPLOY_PATH/$BINARY_NAME"

# Make binary executable
ssh "$LIBVIRT_HOST" "chmod +x $DEPLOY_PATH/$BINARY_NAME"

echo "✓ Deployment complete"
echo ""

# Run phase
echo "=== Starting runner-win on $LIBVIRT_HOST ==="
echo ""
echo "Runner will connect to API at localhost:$LOCAL_API_PORT (via SSH tunnel)"
echo "Runner API will be available at localhost:$RUNNER_PORT (via SSH tunnel)"
echo ""
echo "Press Ctrl+C to stop the runner"
echo "=========================================="
echo ""

# Execute runner in foreground with environment variables
# The runner expects:
# - DAYTONA_API_URL: Points to localhost:3000 which tunnels to local API
# - DAYTONA_RUNNER_TOKEN: Authentication token
# - API_PORT: The port runner listens on
# - LIBVIRT_URI: Connection to local libvirt
# - ENVIRONMENT: development mode
# - LOG_LEVEL: debug for development
ssh -t "$LIBVIRT_HOST" "
    export DAYTONA_API_URL='http://localhost:$LOCAL_API_PORT/api'
    export DAYTONA_RUNNER_TOKEN='$DAYTONA_RUNNER_TOKEN'
    export API_TOKEN='$DAYTONA_RUNNER_TOKEN'
    export API_PORT='$RUNNER_PORT'
    export LIBVIRT_URI='qemu:///system'
    export ENVIRONMENT='development'
    export LOG_LEVEL='debug'
    export RESOURCE_LIMITS_DISABLED='true'
    cd $DEPLOY_PATH && ./$BINARY_NAME
"

