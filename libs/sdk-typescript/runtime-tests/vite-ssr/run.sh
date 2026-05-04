#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
rm -rf node_modules package-lock.json dist
npm install --silent
npm install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"
npx vite build --ssr src/ssr.ts >/dev/null
node --input-type=module -e "
import('./dist/ssr.js').then(async m => {
  const result = await m.run()
  if (result !== 'PASS') throw new Error('Unexpected result: ' + result)
  console.log('PASS')
}).catch(e => { console.error('FAIL:', e.message); process.exit(1) })
"
