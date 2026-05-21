#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0
#
# Reproduces and regression-tests issue #4771: sandbox.fs.downloadFile() in
# Next.js App Router when the SDK is externalized via `serverExternalPackages`.

set -euo pipefail
rm -rf node_modules package-lock.json .next

npm install --silent
npm install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"

# ---------------------------------------------------------------------------
# Create a sandbox and upload a test file for the downloadFile assertion.
# Uses the SDK from Node.js directly (CJS path — not the bug scenario).
# ---------------------------------------------------------------------------
FILE_CONTENT="hello from nextjs-external"
FILE_PATH="test.txt"

# Register cleanup BEFORE creating the sandbox so any failure during creation
# (or between creation and the rest of the script) is still cleaned up.
SANDBOX_ID=""
SERVER_PID=""
cleanup() {
  if [ -n "$SANDBOX_ID" ]; then
    node --input-type=module -e "
      import { Daytona } from '@daytona/sdk';
      const d = new Daytona();
      try { const s = await d.get('${SANDBOX_ID}'); await d.delete(s); } catch {}
    " 2>/dev/null || true
  fi
  if [ -n "$SERVER_PID" ]; then
    kill -9 "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

SANDBOX_ID=$(node --input-type=module -e "
import { Daytona } from '@daytona/sdk';
const d = new Daytona();
const s = await d.create({ timeout: 120, labels: { purpose: 'runtime-test-nextjs-external' } });
await s.fs.uploadFile(Buffer.from('${FILE_CONTENT}'), '${FILE_PATH}');
process.stdout.write(s.id);
")
echo "Sandbox created: $SANDBOX_ID"

NODE_ENV=production npm run build >/dev/null

PORT=${RUNTIME_TEST_PORT:-3804}
NODE_ENV=production npx next start -p "$PORT" >/tmp/nextjs-external-runtime.log 2>&1 &
SERVER_PID=$!

for i in $(seq 1 30); do
  if curl -sf "http://localhost:$PORT/api/sandboxes" >/dev/null 2>&1; then break; fi
  sleep 1
done

# --- Test 1: basic import (Image + list) ---
RESPONSE=$(curl -sf -m 30 "http://localhost:$PORT/api/sandboxes")
echo "Response: $RESPONSE"
echo "$RESPONSE" | grep -q '"imageOk":true' || { echo "FAIL: imageOk false"; exit 1; }
echo "$RESPONSE" | grep -q '"listOk":true' || { echo "FAIL: listOk false"; exit 1; }

# --- Test 2: downloadFile (the bug path from #4771) ---
# urlencode FILE_CONTENT via node to avoid a python3 dependency.
ENCODED_CONTENT=$(node -e "process.stdout.write(encodeURIComponent('${FILE_CONTENT}'))")
DL_RESPONSE=$(curl -sf -m 60 "http://localhost:$PORT/api/download?sandboxId=${SANDBOX_ID}&expected=${ENCODED_CONTENT}")
echo "Download response: $DL_RESPONSE"
echo "$DL_RESPONSE" | grep -q '"downloadOk":true' || { echo "FAIL: downloadOk false"; exit 1; }

echo "PASS"
