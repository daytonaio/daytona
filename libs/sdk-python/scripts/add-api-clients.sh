#!/usr/bin/env bash
set -e

# Load environment variables from .env.local, if it exists
if [ -f "$(dirname "$0")/../.env.local" ]; then
  # shellcheck disable=SC1091
  source "$(dirname "$0")/../.env.local"
fi

echo "→ add-api-clients"

if [ -n "$PY_API_CLIENT_VERSION" ] || [ -n "$PY_PACKAGE_VERSION" ]; then
  echo "Adding API clients at version ${PY_API_CLIENT_VERSION:-$PY_PACKAGE_VERSION}"
  poetry add \
    "daytona_api_client@^${PY_API_CLIENT_VERSION:-$PY_PACKAGE_VERSION}" \
    "daytona_api_client_async@^${PY_API_CLIENT_VERSION:-$PY_PACKAGE_VERSION}"
else
  echo "No override found; skipping add-api-clients"
fi