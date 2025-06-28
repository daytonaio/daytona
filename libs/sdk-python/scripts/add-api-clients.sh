#!/usr/bin/env bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -e

echo "→ add-api-clients"

if [ -n "$PYPI_PKG_VERSION" ]; then
  echo "Adding API clients at version $PYPI_PKG_VERSION"
  poetry add \
    "daytona_api_client@$PYPI_PKG_VERSION" \
    "daytona_api_client_async@$PYPI_PKG_VERSION"
else
  echo "PYPI_PKG_VERSION not set; skipping add-api-clients"
fi
