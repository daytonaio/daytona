#!/usr/bin/env bash
set -euo pipefail

# Adds dynamic version (go:embed) and custom UserAgent to generated Go API clients.
# Usage: postprocess.sh <project-root> <package-name> <client-name>

if [ $# -lt 3 ]; then
  echo "Usage: $0 <project-root> <package-name> <client-name>" >&2
  exit 1
fi

PROJECT_ROOT="$1"
PACKAGE_NAME="$2"
CLIENT_NAME="$3"

cat > "$PROJECT_ROOT/version.go" << EOF
package ${PACKAGE_NAME}

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var _clientVersion string

var ClientVersion = strings.TrimSpace(_clientVersion)
EOF

grep -q 'UserAgent:.*"[^"]*"' "$PROJECT_ROOT/configuration.go" || { echo "ERROR: UserAgent string not found in configuration.go" >&2; exit 1; }
sed -i "s|UserAgent: *\"[^\"]*\"|UserAgent:        \"${CLIENT_NAME}/\" + ClientVersion|" "$PROJECT_ROOT/configuration.go"

echo "Postprocessed Go client at $PROJECT_ROOT"
