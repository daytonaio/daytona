#!/usr/bin/env bash
set -euo pipefail

# Adds custom User-Agent to generated TypeScript API clients using package.json version.
# Usage: postprocess.sh <src-dir> <client-name>

if [ $# -lt 2 ]; then
  echo "Usage: $0 <src-dir> <client-name>" >&2
  exit 1
fi

SRC_DIR="$1"
CLIENT_NAME="$2"
CONFIG="$SRC_DIR/configuration.ts"

sed -i '/Do not edit the class manually/,/\*\//{
  /\*\//a\
\
import * as packageJson from '\''../package.json'\'';
}' "$CONFIG"

if grep -q "'User-Agent'" "$CONFIG"; then
  sed -i "s|'User-Agent': \`[^']*\`|'User-Agent': \`${CLIENT_NAME}/\${packageJson.version}\`|" "$CONFIG"
else
  sed -i "s|\.\.\.param\.baseOptions?.headers,|'User-Agent': \`${CLIENT_NAME}/\${packageJson.version}\`,\n                ...param.baseOptions?.headers,|" "$CONFIG"
fi

echo "Postprocessed TypeScript client at $SRC_DIR"
