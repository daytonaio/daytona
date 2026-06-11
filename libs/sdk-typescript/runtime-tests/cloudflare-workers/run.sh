#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
rm -rf node_modules package-lock.json dist .wrangler
npm install --silent
npm install --silent "$API_CLIENT_TARBALL" "$TOOLBOX_API_CLIENT_TARBALL" "$SDK_TARBALL"

PORT=${RUNTIME_TEST_PORT:-3804}
LOG=/tmp/wrangler-runtime.log

cat > .dev.vars <<EOF
DAYTONA_API_KEY=$DAYTONA_API_KEY
DAYTONA_API_URL=$DAYTONA_API_URL
EOF

PID=""

stop_wrangler() {
  [ -n "$PID" ] && kill -9 "$PID" 2>/dev/null || true
  # `npx wrangler dev` spawns `node .../wrangler-dist/cli.js dev`, which the
  # 'wrangler dev' pattern does not match — kill it explicitly or it leaks
  # and keeps holding the port.
  pkill -9 -f 'wrangler dev' 2>/dev/null || true
  pkill -9 -f 'wrangler-dist/cli.js' 2>/dev/null || true
  pkill -9 -f workerd 2>/dev/null || true
}

cleanup() {
  stop_wrangler
  rm -f .dev.vars
}
trap cleanup EXIT

start_wrangler() {
  # --inspector-port 0 picks a free port, avoiding collisions on the fixed default (9229)
  npx wrangler dev --local --port "$PORT" --inspector-port 0 >"$LOG" 2>&1 &
  PID=$!
}

dump_log() {
  echo "--- wrangler log (last 100 lines) ---" >&2
  tail -100 "$LOG" >&2 || true
  echo "--- end wrangler log ---" >&2
}

# Waits until the worker responds 2xx. Bails out early if wrangler dies.
probe() {
  for _ in $(seq 1 30); do
    if ! kill -0 "$PID" 2>/dev/null; then
      echo "wrangler process exited prematurely" >&2
      return 1
    fi
    if curl -sf "http://localhost:$PORT/" >/dev/null 2>&1; then return 0; fi
    sleep 1
  done
  echo "worker did not respond successfully within 30s" >&2
  return 1
}

start_wrangler
if ! probe; then
  dump_log
  echo "Retrying once with a fresh wrangler process..." >&2
  stop_wrangler
  sleep 2
  start_wrangler
  if ! probe; then
    dump_log
    echo "FAIL: worker never served a successful response"
    exit 1
  fi
fi

RESPONSE=$(curl -sf -m 10 "http://localhost:$PORT/") || { dump_log; echo "FAIL: request to worker failed"; exit 1; }
echo "Response: $RESPONSE"

echo "$RESPONSE" | grep -q '"imageOk":true' || { echo "FAIL: imageOk false"; exit 1; }
echo "$RESPONSE" | grep -q '"listOk":true' || { echo "FAIL: listOk false"; exit 1; }
echo "PASS"
