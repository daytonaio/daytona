#!/bin/bash
# Increment build number and output it

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_FILE="$SCRIPT_DIR/../BUILD_NUMBER"

# Read current build number (default to 0 if file doesn't exist)
if [ -f "$BUILD_FILE" ]; then
    BUILD_NUM=$(cat "$BUILD_FILE")
else
    BUILD_NUM=0
fi

# Increment
BUILD_NUM=$((BUILD_NUM + 1))

# Write back
echo "$BUILD_NUM" > "$BUILD_FILE"

# Output for use in build
echo "$BUILD_NUM"

