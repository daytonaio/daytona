#!/bin/bash
set -e  # Exit immediately if any command fails

# Check if there are any staged changes in the _async or _sync folders
ASYNC_CHANGES=$(git diff --cached --name-only | grep "libs/sdk-python/src/daytona_sdk/_async/.*\.py$")
SYNC_CHANGES=$(git diff --cached --name-only | grep "libs/sdk-python/src/daytona_sdk/_sync/.*\.py$")

if [ -n "$ASYNC_CHANGES" ] || [ -n "$SYNC_CHANGES" ]; then
    if [ -n "$ASYNC_CHANGES" ] && [ -n "$SYNC_CHANGES" ]; then
        echo "Detected changes in both _async and _sync folders. Running sync generator..."
    elif [ -n "$ASYNC_CHANGES" ]; then
        echo "Detected changes in _async folder. Running sync generator..."
    else
        echo "Detected changes in _sync folder. Running sync generator to ensure consistency..."
    fi
    
    # Run the sync generator - will automatically fail the hook if this fails
    yarn sdk-python:generate-sync
    git add libs/sdk-python/src/daytona_sdk/_sync/
else
    echo "No changes detected in _async or _sync folders. Skipping sync generation."
fi 