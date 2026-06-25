#!/usr/bin/env bash
#
# browsecli-demo.sh — the load-bearing demo every sandbox template runs.
#
# Inside ANY sandbox (E2B, Modal, Daytona, Vercel, Cloudflare, Fly, ...), this
# script uses the `browse` CLI to drive a *remote* Browserbase session. The
# browser never runs in the sandbox — the sandbox runs the agent/CLI and
# connects out over CDP to a Verified Browserbase browser that:
#   - uses a residential/verified IP (no datacenter-IP blocking)
#   - runs in Verified browser mode (passes bot-detection fingerprinting)
#   - auto-solves CAPTCHAs / challenges server-side
#
# Requires: BROWSERBASE_API_KEY in env.
# Optional: TARGET_URL (default https://nowsecure.nl, a Cloudflare-protected page)
#
set -euo pipefail

TARGET="${TARGET_URL:-https://nowsecure.nl}"
: "${BROWSERBASE_API_KEY:?BROWSERBASE_API_KEY must be set}"

log() { printf '[browsecli-demo] %s\n' "$*"; }

# Parse a top-level JSON string field without requiring jq (node is always present).
jget() { node -e 'let d="";process.stdin.on("data",c=>d+=c).on("end",()=>{try{process.stdout.write(String(JSON.parse(d)["'"$1"'"]??""))}catch(e){process.exit(2)}})'; }

log "browse version: $(browse --version 2>/dev/null || echo unknown)"
log "creating Verified Browserbase session (proxies + verified + solve-captchas)..."
SJSON="$(browse cloud sessions create --proxies --verified --solve-captchas --keep-alive --timeout 300)"
CONNECT="$(printf '%s' "$SJSON" | jget connectUrl)"
SID="$(printf '%s' "$SJSON" | jget id)"
[ -n "$CONNECT" ] || { log "FAIL: no connectUrl returned (check BROWSERBASE_API_KEY)"; exit 1; }
log "session ready: ${SID}"

cleanup() { browse stop --session demo >/dev/null 2>&1 || true; }
trap cleanup EXIT

log "opening protected target: ${TARGET}"
browse open "$TARGET" --cdp "$CONNECT" --session demo --wait load --timeout 60000 >/dev/null
sleep 3

TITLE="$(browse get title --session demo | jget title)"
BODY="$(browse get text body --session demo | jget text)"
BLEN="${#BODY}"

log "page title : ${TITLE:-<empty>}"
log "body length: ${BLEN} chars"
printf '%s\n' "----- first 400 chars of page text -----"
printf '%s\n' "${BODY:0:400}"
printf '%s\n' "----------------------------------------"

# Heuristic pass/fail: a challenge wall returns an empty/short body or a
# "checking your browser" interstitial; a solved page returns real content.
shopt -s nocasematch
if [ "$BLEN" -lt 50 ] || [[ "$BODY" == *"just a moment"* ]] || [[ "$BODY" == *"checking your browser"* ]] || [[ "$BODY" == *"enable javascript and cookies"* ]]; then
  log "RESULT: ❌ BLOCKED — looks like a challenge wall, not real content"
  exit 1
fi
log "RESULT: ✅ PASS — reached real content through the protected site from inside the sandbox"
