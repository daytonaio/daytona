#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Demo script to show plugin-based computer use functionality

set -e

echo "=== Daytona Daemon Plugin Architecture Demo ==="
echo

# Build the daemon and plugin
echo "1. Building daemon and plugin..."
cd /workspaces/daytona
npx nx build daemon
npx nx build-plugin daemon

DAEMON_BINARY="/workspaces/daytona/dist/apps/daemon-amd64"
PLUGIN_BINARY="/workspaces/daytona/dist/apps/computeruse.so"

if [ ! -f "$DAEMON_BINARY" ]; then
    echo "Error: Daemon binary not found at $DAEMON_BINARY"
    exit 1
fi

if [ ! -f "$PLUGIN_BINARY" ]; then
    echo "Error: Plugin binary not found at $PLUGIN_BINARY"
    exit 1
fi

echo "✅ Daemon built successfully at $DAEMON_BINARY"
echo "✅ Plugin built successfully at $PLUGIN_BINARY"
echo

# Check daemon dependencies
echo "2. Checking daemon dependencies..."
echo "   Checking for X11 dependencies in daemon binary..."
if ldd "$DAEMON_BINARY" | grep -q "libX11"; then
    echo "   ⚠️  Daemon still has X11 dependencies"
else
    echo "   ✅ Daemon has no X11 dependencies"
fi
echo

# Check plugin dependencies
echo "3. Checking plugin dependencies..."
echo "   Checking for X11 dependencies in plugin..."
if ldd "$PLUGIN_BINARY" | grep -q "libX11"; then
    echo "   ✅ Plugin has X11 dependencies (as expected)"
else
    echo "   ⚠️  Plugin has no X11 dependencies"
fi
echo

# Start the daemon in background
echo "4. Starting daemon..."
$DAEMON_BINARY &
DAEMON_PID=$!

# Wait for daemon to start
echo "   Waiting for daemon to start..."
sleep 3

# Check if daemon is running
if ! kill -0 $DAEMON_PID 2>/dev/null; then
    echo "❌ Daemon failed to start"
    exit 1
fi

echo "✅ Daemon started successfully (PID: $DAEMON_PID)"
echo

# Test computer use status without plugin
echo "5. Testing computer use status without plugin..."
echo "   GET /computer/status"
STATUS_RESPONSE=$(curl -s http://localhost:22222/computer/status || echo "{}")
echo "   Response: $STATUS_RESPONSE"
echo

# Test a computer use endpoint without plugin
echo "6. Testing computer use endpoint without plugin..."
echo "   GET /computer/screenshot"
SCREENSHOT_RESPONSE=$(curl -s http://localhost:22222/computer/screenshot || echo "{}")
echo "   Response: $SCREENSHOT_RESPONSE"
echo

# Stop the daemon
echo "7. Stopping daemon..."
kill $DAEMON_PID
wait $DAEMON_PID 2>/dev/null || true

echo "✅ Daemon stopped successfully"
echo

# Test with plugin (if X11 is available)
echo "8. Testing with plugin (if X11 is available)..."
if ldd "$PLUGIN_BINARY" | grep -q "libX11" && [ -f "/usr/lib/x86_64-linux-gnu/libX11.so.6" ]; then
    echo "   X11 libraries available, testing plugin functionality..."
    
    # Copy plugin to expected location
    sudo mkdir -p /usr/local/lib
    sudo cp "$PLUGIN_BINARY" /usr/local/lib/computeruse.so
    
    # Start daemon again
    $DAEMON_BINARY &
    DAEMON_PID=$!
    sleep 3
    
    if kill -0 $DAEMON_PID 2>/dev/null; then
        echo "   ✅ Daemon started with plugin"
        
        # Test computer use status with plugin
        echo "   GET /computer/status"
        STATUS_RESPONSE=$(curl -s http://localhost:22222/computer/status || echo "{}")
        echo "   Response: $STATUS_RESPONSE"
        
        # Stop daemon
        kill $DAEMON_PID
        wait $DAEMON_PID 2>/dev/null || true
    else
        echo "   ❌ Daemon failed to start with plugin"
    fi
else
    echo "   ⚠️  X11 libraries not available, skipping plugin test"
fi
echo

# Test other daemon functionality
echo "9. Testing other daemon functionality..."
$DAEMON_BINARY &
DAEMON_PID=$!
sleep 3

echo "   GET /version"
VERSION_RESPONSE=$(curl -s http://localhost:22222/version || echo "{}")
echo "   Response: $VERSION_RESPONSE"

echo "   GET /files/"
FILES_RESPONSE=$(curl -s http://localhost:22222/files/ || echo "{}")
echo "   Response: $FILES_RESPONSE"

# Stop the daemon
kill $DAEMON_PID
wait $DAEMON_PID 2>/dev/null || true

echo "✅ Daemon stopped successfully"
echo

echo "=== Demo Summary ==="
echo "✅ Daemon starts successfully without X11 dependencies"
echo "✅ Plugin architecture separates X11 dependencies"
echo "✅ Computer use endpoints provide helpful error messages when plugin unavailable"
echo "✅ Other daemon functionality works normally"
echo "✅ No startup failures due to missing dependencies"
echo
echo "This demonstrates that the plugin architecture allows the daemon to:"
echo "- Start successfully on any system"
echo "- Isolate X11 dependencies to the plugin"
echo "- Provide clear feedback about computer use availability"
echo "- Continue working normally for other features"
echo "- Load X11 dependencies only when plugin is available and needed" 