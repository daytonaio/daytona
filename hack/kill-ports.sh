#!/usr/bin/env bash
#
# Kill all processes running on Daytona application ports.
#
set -euo pipefail

if [ "$EUID" -ne 0 ]; then
  echo "Error: this script must be run with sudo (runner runs as root)." >&2
  exit 1
fi

for cmd in lsof ps; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "Error: '$cmd' is not installed. Please install it and try again." >&2
    exit 1
  fi
done

PORTS=(
  3000  # dashboard
  3001  # api
  3003  # runner
  3009  # cli (auth0 callback)
  4000  # proxy
  4321  # docs (functions)
)

echo "Daytona application ports: ${PORTS[*]}"
echo ""

declare -A SEEN_PIDS
KILLED=0
FAILED=0

for PORT in "${PORTS[@]}"; do
  PIDS=$(lsof -ti ":$PORT" 2>/dev/null || true)
  if [ -n "$PIDS" ]; then
    for PID in $PIDS; do
      if [[ -n "${SEEN_PIDS[$PID]+x}" ]]; then
        continue
      fi
      SEEN_PIDS[$PID]=1

      PROC_NAME=$(ps -p "$PID" -o comm= 2>/dev/null || echo "unknown")
      if kill -9 "$PID" 2>/dev/null; then
        echo "Killed PID $PID ($PROC_NAME) on port $PORT"
        KILLED=$((KILLED + 1))
      else
        echo "Failed to kill PID $PID ($PROC_NAME) on port $PORT — permission denied?" >&2
        FAILED=$((FAILED + 1))
      fi
    done
  fi
done

echo ""
if [ "$KILLED" -eq 0 ] && [ "$FAILED" -eq 0 ]; then
  echo "No processes found on any application ports."
else
  [ "$KILLED" -gt 0 ] && echo "Killed $KILLED process(es)."
  [ "$FAILED" -gt 0 ] && echo "Failed to kill $FAILED process(es)." >&2
  [ "$FAILED" -gt 0 ] && exit 1
fi
