#!/usr/bin/env bash
set -e

# Load environment variables from .env.local, if it exists
if [ -f "$(dirname "$0")/../.env.local" ]; then
  # shellcheck disable=SC1091
  source "$(dirname "$0")/../.env.local"
fi

echo "→ build-sdk"

if [ -n "$PY_PACKAGE_VERSION" ]; then
  echo "Bumping SDK version to $PY_PACKAGE_VERSION"
  poetry version "$PY_PACKAGE_VERSION"
else
  echo "Using version from pyproject.toml"
fi

poetry build
