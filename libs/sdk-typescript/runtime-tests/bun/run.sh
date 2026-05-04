#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if ! command -v bun >/dev/null 2>&1; then
  echo "SKIP: bun not installed"
  exit 0
fi

rm -rf node_modules package-lock.json bun.lock bun.lockb
node -e "
const fs = require('fs');
const p = require('./package.json');
p.overrides = {
  '@daytona/api-client': '$API_CLIENT_TARBALL',
  '@daytona/toolbox-api-client': '$TOOLBOX_API_CLIENT_TARBALL',
};
fs.writeFileSync('./package.json', JSON.stringify(p, null, 2));
"
bun install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"
bun run test.ts
