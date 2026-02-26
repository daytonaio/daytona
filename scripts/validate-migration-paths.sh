#!/bin/bash

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Fails if any migration file is placed directly under
# apps/api/src/migrations/ instead of pre-deploy/ or post-deploy/.
# Legacy migrations (timestamp <= LEGACY_CUTOFF) are excluded.

LEGACY_CUTOFF=1770880371265

forbidden=()

for f in "$@"; do
  rel="$(realpath --relative-to="$PWD" "$f")"

  case "$rel" in
    apps/api/src/migrations/pre-deploy/*.ts) ;;
    apps/api/src/migrations/post-deploy/*.ts) ;;
    apps/api/src/migrations/*.ts)
      timestamp=$(basename "$rel" | grep -oP '^\d+')
      if [ -n "$timestamp" ] && [ "$timestamp" -le "$LEGACY_CUTOFF" ]; then
        continue
      fi
      forbidden+=("$rel")
      ;;
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
