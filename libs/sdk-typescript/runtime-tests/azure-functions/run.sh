#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if ! command -v func >/dev/null 2>&1; then
  echo "SKIP: azure-functions-core-tools not installed (install: npm i -g azure-functions-core-tools@4)"
  exit 0
fi

rm -rf node_modules package-lock.json dist
npm install --silent
npm install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"
npm run build >/dev/null

cat > local.settings.json <<EOF
{
  "IsEncrypted": false,
  "Values": {
    "AzureWebJobsStorage": "",
    "FUNCTIONS_WORKER_RUNTIME": "node",
    "DAYTONA_API_KEY": "$DAYTONA_API_KEY",
    "DAYTONA_API_URL": "$DAYTONA_API_URL"
  }
}
EOF

PORT=${RUNTIME_TEST_PORT:-3805}

env DAYTONA_API_KEY="$DAYTONA_API_KEY" DAYTONA_API_URL="$DAYTONA_API_URL" \
  func start --port "$PORT" >/tmp/azure-runtime.log 2>&1 &
PID=$!
trap "kill -9 $PID 2>/dev/null || true; pkill -9 -f 'func.*start' 2>/dev/null || true" EXIT

for i in $(seq 1 60); do
  if curl -sf "http://localhost:$PORT/api/sandboxes" >/dev/null 2>&1; then break; fi
  sleep 1
done

RESPONSE=$(curl -sf -m 10 "http://localhost:$PORT/api/sandboxes")
echo "Response: $RESPONSE"

echo "$RESPONSE" | grep -q '"imageOk":true' || { echo "FAIL: imageOk not true"; exit 1; }
echo "$RESPONSE" | grep -q '"listOk":true' || { echo "FAIL: listOk not true"; exit 1; }
echo "PASS"
