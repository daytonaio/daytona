#!/bin/bash

# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

set -e  # Exit immediately if any command fails

# Check if there are any staged changes in the _async or _sync folders
ASYNC_CHANGES=$(git diff --cached --name-only | grep "libs/sdk-python/src/daytona/_async/.*\.py$" || true)
SYNC_CHANGES=$(git diff --cached --name-only | grep "libs/sdk-python/src/daytona/_sync/.*\.py$" || true)

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
    isort libs/sdk-python/src/daytona/_sync
    black libs/sdk-python/src/daytona/_sync --config pyproject.toml
    
    # Check if there are any new changes after running the sync generator
    NEW_CHANGES=$(git diff --name-only libs/sdk-python/src/daytona/_sync)
    
    if [ -n "$NEW_CHANGES" ]; then
        echo "The sync generator has created new changes in the _sync folder:"
        echo "$NEW_CHANGES"
        echo ""
        echo "Please review these changes and add them to your commit:"
        echo "  git add libs/sdk-python/src/daytona/_sync/"
        echo ""
        echo "Then retry your commit."
        exit 1
    else
        echo "Sync generator completed with no new changes."
    fi
else
    echo "No changes detected in _async or _sync folders. Skipping sync generation."
fi 