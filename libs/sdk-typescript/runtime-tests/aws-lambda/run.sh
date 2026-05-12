#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
rm -rf node_modules package-lock.json dist
npm install --silent
npm install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"
npm run build >/dev/null

RESPONSE=$(node -e "
const lambdaLocal = require('lambda-local');
const handler = require('./dist/handler.js').handler;
lambdaLocal.execute({
  lambdaFunc: { handler },
  lambdaHandler: 'handler',
  event: {},
  environment: { DAYTONA_API_KEY: process.env.DAYTONA_API_KEY, DAYTONA_API_URL: process.env.DAYTONA_API_URL },
  timeoutMs: 30000,
  verboseLevel: 0,
}).then(r => process.stdout.write(r.body)).catch(e => { console.error('LAMBDA ERROR:', e.message); process.exit(1) });
")
echo "Response: $RESPONSE"

echo "$RESPONSE" | grep -q '"imageOk":true' || { echo "FAIL: imageOk not true"; exit 1; }
echo "$RESPONSE" | grep -q '"listOk":true' || { echo "FAIL: listOk not true"; exit 1; }
echo "PASS"
