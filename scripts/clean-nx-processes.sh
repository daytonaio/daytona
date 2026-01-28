#!/bin/bash

# Script to clean up all stale nx processes
# Run this when nx processes get stuck or before starting fresh with yarn serve

echo "🧹 Cleaning up nx processes..."

# Kill nx main processes
pkill -9 -f "nx/bin/nx.js" 2>/dev/null && echo "  ✓ Killed nx.js processes" || echo "  - No nx.js processes found"

# Kill nx run executors
pkill -9 -f "nx/bin/run-executor" 2>/dev/null && echo "  ✓ Killed run-executor processes" || echo "  - No run-executor processes found"

# Kill nx plugin workers
pkill -9 -f "nx/src/project-graph/plugins/isolation/plugin-worker" 2>/dev/null && echo "  ✓ Killed plugin-worker processes" || echo "  - No plugin-worker processes found"

# Kill node-with-require-overrides (nx js executor)
pkill -9 -f "node-with-require-overrides" 2>/dev/null && echo "  ✓ Killed node-with-require-overrides processes" || echo "  - No node-with-require-overrides processes found"

# Kill vite dev servers
pkill -9 -f "node.*vite" 2>/dev/null && echo "  ✓ Killed vite processes" || echo "  - No vite processes found"

# Kill astro dev servers
pkill -9 -f "astro dev" 2>/dev/null && echo "  ✓ Killed astro processes" || echo "  - No astro processes found"

# Kill gow (go watch) processes
pkill -9 -f "gow run" 2>/dev/null && echo "  ✓ Killed gow processes" || echo "  - No gow processes found"

# Wait a moment for processes to terminate
sleep 1

# Show remaining node processes (for verification)
remaining=$(pgrep -c -f "nx" 2>/dev/null || echo "0")
if [ "$remaining" -gt 0 ]; then
    echo ""
    echo "⚠️  $remaining nx-related processes may still be running"
    echo "   Run 'ps aux | grep nx' to check"
else
    echo ""
    echo "✅ All nx processes cleaned up!"
fi
