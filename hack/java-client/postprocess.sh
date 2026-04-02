#!/usr/bin/env bash
set -euo pipefail

# Sets custom User-Agent in generated Java API clients using the project version.
# Usage: postprocess.sh <project-root> <client-name>

if [ $# -lt 2 ]; then
  echo "Usage: $0 <project-root> <client-name>" >&2
  exit 1
fi

PROJECT_ROOT="$1"
CLIENT_NAME="$2"
API_CLIENT=$(find "$PROJECT_ROOT/src" -name "ApiClient.java" -path "*/client/ApiClient.java" | head -1)

if [ -z "$API_CLIENT" ] || [ ! -f "$API_CLIENT" ]; then
  echo "ERROR: ApiClient.java not found in $PROJECT_ROOT" >&2
  exit 1
fi

VERSION=$(grep "^version" "$PROJECT_ROOT/build.gradle" | sed "s/version = '//" | sed "s/'//")
sed -i "s|setUserAgent(\"OpenAPI-Generator/[^\"]*\")|setUserAgent(\"${CLIENT_NAME}/${VERSION}\")|" "$API_CLIENT"

echo "Postprocessed Java client at $PROJECT_ROOT (User-Agent: ${CLIENT_NAME}/${VERSION})"
