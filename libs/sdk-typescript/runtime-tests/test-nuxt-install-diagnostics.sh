#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

FAKE_BIN="$TMP_DIR/bin"
mkdir -p "$FAKE_BIN"

cat >"$FAKE_BIN/npm" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [ "${1:-}" != "install" ]; then
  echo "unexpected npm command: $*" >&2
  exit 99
fi

COUNT_FILE="${NPM_CALL_COUNT_FILE:?}"
count=0
if [ -f "$COUNT_FILE" ]; then
  count="$(cat "$COUNT_FILE")"
fi
count=$((count + 1))
echo "$count" >"$COUNT_FILE"

if [ "$count" -lt "${NPM_FAIL_ON_CALL:?}" ]; then
  exit 0
fi

for arg in "$@"; do
  if [ "$arg" = "--silent" ]; then
    exit 42
  fi
done

echo "npm failure detail: registry unavailable" >&2
exit 42
EOF
chmod +x "$FAKE_BIN/npm"

assert_install_failure_is_visible() {
  local fail_on_call="$1"
  local work_dir="$TMP_DIR/nuxt-$fail_on_call"
  local count_file="$TMP_DIR/npm-count-$fail_on_call"

  cp -R "$ROOT/nuxt" "$work_dir"

  set +e
  OUTPUT="$(
    cd "$work_dir" &&
      PATH="$FAKE_BIN:$PATH" \
        NPM_CALL_COUNT_FILE="$count_file" \
        NPM_FAIL_ON_CALL="$fail_on_call" \
        API_CLIENT_TARBALL="$TMP_DIR/api-client.tgz" \
        TOOLBOX_API_CLIENT_TARBALL="$TMP_DIR/toolbox-api-client.tgz" \
        SDK_TARBALL="$TMP_DIR/sdk.tgz" \
        bash run.sh 2>&1
  )"
  STATUS=$?
  set -e

  if [ "$STATUS" -eq 0 ]; then
    echo "expected nuxt runtime script to fail when npm install fails" >&2
    exit 1
  fi

  if ! grep -q "npm failure detail: registry unavailable" <<<"$OUTPUT"; then
    echo "expected npm failure details in nuxt install output" >&2
    echo "$OUTPUT" >&2
    exit 1
  fi

  if grep -q "unexpected npm command" <<<"$OUTPUT"; then
    echo "$OUTPUT" >&2
    exit 1
  fi
}

assert_install_failure_is_visible 1
assert_install_failure_is_visible 2
