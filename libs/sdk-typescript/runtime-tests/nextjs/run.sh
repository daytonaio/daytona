#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
rm -rf node_modules package-lock.json .next
npm install --silent
npm install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"
npm run build >/dev/null

PORT=${RUNTIME_TEST_PORT:-3801}
npx next start -p "$PORT" >/tmp/nextjs-runtime.log 2>&1 &
PID=$!
trap "kill -9 $PID 2>/dev/null || true; pkill -9 -f 'next start' 2>/dev/null || true" EXIT

for i in $(seq 1 30); do
  if curl -sf "http://localhost:$PORT/api/sandboxes" >/dev/null 2>&1; then break; fi
  sleep 1
done

RESPONSE=$(curl -sf -m 10 "http://localhost:$PORT/api/sandboxes")
echo "Response: $RESPONSE"

echo "$RESPONSE" | grep -q '"imageOk":true' || { echo "FAIL: imageOk false"; exit 1; }
echo "$RESPONSE" | grep -q '"listOk":true' || { echo "FAIL: listOk false"; exit 1; }
echo "PASS"
