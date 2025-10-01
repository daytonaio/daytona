#!/bin/sh

set -e

cd "$(dirname "$0")"

cat <<EOF > src/index.ts
export * as v1 from './proto/runner/v1/runner'
export * as v2 from './proto/runner/v2/runner'
EOF