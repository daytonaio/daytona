#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Establish SSH tunnels for runner-win communication
# This creates bidirectional tunnels so the local API and remote runner
# can communicate as if they were on the same machine (localhost)
#
# Tunnels:
#   Local :8080  -> Remote :8080  (API can reach runner at localhost:8080)
#   Remote :3000 -> Local :3000   (Runner can reach API at localhost:3000)
#
# Usage:
#   ./tunnel.sh              # Start tunnels (foreground)
#   ./tunnel.sh --background # Start tunnels in background
#
# Environment variables:
#   LIBVIRT_HOST - Remote host (default: h1001.blinkbox.dev)
#   RUNNER_PORT  - Runner API port (default: 8080)
#   API_PORT     - Local API port (default: 3000)

set -e

# Configuration with defaults
LIBVIRT_HOST="${LIBVIRT_HOST:-h1001.blinkbox.dev}"
RUNNER_PORT="${RUNNER_PORT:-8080}"
API_PORT="${API_PORT:-3001}"

echo "=== Daytona Runner-Win SSH Tunnels ==="
echo "Libvirt Host: $LIBVIRT_HOST"
echo ""
echo "Tunnels to establish:"
echo "  Local :$RUNNER_PORT  -> Remote :$RUNNER_PORT  (API -> Runner)"
echo "  Remote :$API_PORT -> Local :$API_PORT   (Runner -> API)"
echo ""

# Check if tunnels are already running
check_existing_tunnel() {
    if pgrep -f "ssh.*-L $RUNNER_PORT:localhost:$RUNNER_PORT.*$LIBVIRT_HOST" > /dev/null 2>&1; then
        echo "Warning: SSH tunnel to $LIBVIRT_HOST may already be running"
        echo "Check with: ps aux | grep 'ssh.*$LIBVIRT_HOST'"
        echo ""
    fi
}

check_existing_tunnel

# Build SSH command with tunnels
# -L local_port:remote_host:remote_port - Local forward (access remote service locally)
# -R remote_port:local_host:local_port  - Remote forward (access local service from remote)
# -N - Don't execute remote command (tunnel only)
# -o ServerAliveInterval=60 - Send keepalive every 60 seconds
# -o ServerAliveCountMax=3 - Disconnect after 3 missed keepalives
# -o ExitOnForwardFailure=yes - Exit if port forwarding fails

SSH_CMD="ssh \
    -L $RUNNER_PORT:localhost:$RUNNER_PORT \
    -R $API_PORT:localhost:$API_PORT \
    -N \
    -o ServerAliveInterval=60 \
    -o ServerAliveCountMax=3 \
    -o ExitOnForwardFailure=yes \
    $LIBVIRT_HOST"

if [ "$1" = "--background" ] || [ "$1" = "-b" ]; then
    echo "Starting tunnels in background..."
    $SSH_CMD &
    TUNNEL_PID=$!
    echo "✓ Tunnels started (PID: $TUNNEL_PID)"
    echo ""
    echo "To stop tunnels: kill $TUNNEL_PID"
    echo "Or: pkill -f 'ssh.*-L $RUNNER_PORT.*$LIBVIRT_HOST'"
else
    echo "Starting tunnels in foreground..."
    echo "Press Ctrl+C to stop"
    echo "=========================================="
    echo ""
    exec $SSH_CMD
fi


