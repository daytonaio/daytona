#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if ! command -v deno >/dev/null 2>&1; then
  echo "SKIP: deno not installed"
  exit 0
fi

rm -rf node_modules package-lock.json
cat > package.json <<'JSON'
{ "name": "runtime-test-deno", "private": true, "type": "module" }
JSON
npm install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"
deno run -A --no-check test.ts
