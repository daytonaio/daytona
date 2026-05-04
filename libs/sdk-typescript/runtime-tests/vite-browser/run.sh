#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
rm -rf node_modules package-lock.json dist
npm install --silent
npm install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"

if ! npx playwright install --dry-run chromium 2>&1 | grep -q 'is already installed'; then
  npx playwright install --with-deps chromium >/dev/null 2>&1 || npx playwright install chromium >/dev/null 2>&1 || true
fi

npx vite build >/dev/null
node test-browser.mjs
