#!/usr/bin/env bash
#
# Kill all processes running on Daytona application ports.
#
set -euo pipefail

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

KILLED=0

for PORT in "${PORTS[@]}"; do
  PIDS=$(lsof -ti ":$PORT" 2>/dev/null || true)
  if [ -n "$PIDS" ]; then
    for PID in $PIDS; do
      PROC_NAME=$(ps -p "$PID" -o comm= 2>/dev/null || echo "unknown")
      echo "Killing PID $PID ($PROC_NAME) on port $PORT"
      kill -9 "$PID" 2>/dev/null || true
      KILLED=$((KILLED + 1))
    done
  fi
done

if [ "$KILLED" -eq 0 ]; then
  echo "No processes found on any application ports."
else
  echo ""
  echo "Killed $KILLED process(es)."
fi
