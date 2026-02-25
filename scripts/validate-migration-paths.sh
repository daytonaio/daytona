#! /bin/bash

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Fails if any staged migration file is placed directly under
# apps/api/src/migrations/ instead of pre-deploy/ or post-deploy/.

forbidden=()

for f in "$@"; do
  rel="$(realpath --relative-to="$PWD" "$f")"

  case "$rel" in
    apps/api/src/migrations/pre-deploy/*.ts) ;;
    apps/api/src/migrations/post-deploy/*.ts) ;;
    apps/api/src/migrations/*.ts) forbidden+=("$rel") ;;
  esac
done

if [ ${#forbidden[@]} -gt 0 ]; then
  echo "Migration files must be placed in one of:" >&2
  echo "  - apps/api/src/migrations/pre-deploy/" >&2
  echo "  - apps/api/src/migrations/post-deploy/" >&2
  echo "" >&2
  echo "Invalid paths:" >&2
  for p in "${forbidden[@]}"; do
    echo "  - $p" >&2
  done
  echo "" >&2
  echo "See apps/api/src/migrations/README.md for more information." >&2
  exit 1
fi
