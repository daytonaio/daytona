#!/usr/bin/env bash
set -euo pipefail

# Adds dynamic version and custom User-Agent to generated TypeScript API clients.
# Usage: postprocess.sh <src-dir> <client-name>

if [ $# -lt 2 ]; then
  echo "Usage: $0 <src-dir> <client-name>" >&2
  exit 1
fi

SRC_DIR="$1"
CLIENT_NAME="$2"
VERSION="${DEFAULT_PACKAGE_VERSION:-0.0.0-dev}"
CONFIG="$SRC_DIR/configuration.ts"

echo "export const VERSION = '$VERSION';" > "$SRC_DIR/version.ts"

sed -i '/Do not edit the class manually/,/\*\//{
  /\*\//a\
\
import { VERSION } from '\''./version'\'';
}' "$CONFIG"

if grep -q "'User-Agent'" "$CONFIG"; then
  sed -i "s|'User-Agent': \"[^\"]*\"|'User-Agent': \`${CLIENT_NAME}/\${VERSION}\`|" "$CONFIG"
else
  sed -i "s|\.\.\.param\.baseOptions?.headers,|'User-Agent': \`${CLIENT_NAME}/\${VERSION}\`,\n                ...param.baseOptions?.headers,|" "$CONFIG"
fi

echo "Postprocessed TypeScript client at $SRC_DIR"
