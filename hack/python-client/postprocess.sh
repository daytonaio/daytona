#!/usr/bin/env bash
set -euo pipefail

# This script normalizes generated Python client metadata after OpenAPI generation.
# Usage: postprocess.sh <projectRoot>

if [ $# -lt 1 ]; then
  echo "Usage: $0 <projectRoot>" >&2
  exit 1
fi

PROJECT_ROOT="$1"

# Set license in pyproject.toml to Apache-2.0
sed -i 's/^license = ".*"/license = "Apache-2.0"/' "$PROJECT_ROOT/pyproject.toml"

# Ensure urllib3 lower bound is pinned to version 2.1.0 in pyproject.toml, setup.py, and requirements.txt.
# This prevents compatibility issues such as:
# `TypeError: PoolKey.__new__() got an unexpected keyword argument 'key_ca_cert_data'`
# which occur with urllib3 versions earlier than 2.1.0.
sed -i -E 's/(urllib3[^0-9\n]*)([0-9]+\.[0-9]+\.[0-9]+)/\12.1.0/g' \
  "$PROJECT_ROOT/pyproject.toml" \
  "$PROJECT_ROOT/setup.py" \
  "$PROJECT_ROOT/requirements.txt"

# Replace all aliases with serialization_aliases in the models directory so that type checking works.
pkg_root=$(find "$PROJECT_ROOT" -mindepth 1 -maxdepth 2 -type f -name "py.typed" -printf '%h\n' | head -n 1)
MODELS_DIR="$pkg_root/models"
find "$MODELS_DIR" -type f -name "*.py" | while read -r f; do
  sed -i'' -E '/Field\(/ s/alias="([^"]+)"/serialization_alias="\1"/g' "$f"
done

echo "Postprocessed Python client at $PROJECT_ROOT"



