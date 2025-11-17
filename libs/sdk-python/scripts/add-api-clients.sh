#!/usr/bin/env bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -e

echo "→ add-api-clients"

if [ -n "$PYPI_PKG_VERSION" ]; then
  echo "Adding API clients at version $PYPI_PKG_VERSION"

  max_attempts=20
  delay_seconds=5

  last_error=""

  for attempt in $(seq 1 "$max_attempts"); do
    echo "Attempt $attempt/$max_attempts: installing API clients"
    if output=$(poetry add \
      "daytona_api_client@$PYPI_PKG_VERSION" \
      "daytona_api_client_async@$PYPI_PKG_VERSION" \
      "daytona_toolbox_api_client@$PYPI_PKG_VERSION" \
      "daytona_toolbox_api_client_async@$PYPI_PKG_VERSION" 2>&1); then
      echo "Successfully added API clients on attempt $attempt"
      break
    fi

    last_error="$output"

    if [ "$attempt" -lt "$max_attempts" ]; then
      echo "poetry add failed; retrying in ${delay_seconds}s..."
      sleep "$delay_seconds"
    else
      echo "Failed to add API clients after $max_attempts attempts"
      echo "Last error output:" >&2
      echo "$last_error" >&2
      exit 1
    fi
  done
else
  echo "PYPI_PKG_VERSION not set; skipping add-api-clients"
fi
