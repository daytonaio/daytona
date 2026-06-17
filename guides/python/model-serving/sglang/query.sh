#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# ENDPOINT and TOKEN are printed by serve_sglang.py.
: "${ENDPOINT:?set ENDPOINT to the preview URL}"
: "${TOKEN:?set TOKEN to the preview token}"

# max_tokens covers reasoning plus answer; gpt-oss thinks before it speaks
curl -sS --connect-timeout 30 --max-time 120 "$ENDPOINT/v1/chat/completions" \
  -H "x-daytona-preview-token: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-oss-20b",
    "messages": [{"role": "user", "content": "Write a haiku about a sandbox where AI agents run code."}],
    "max_tokens": 4096
  }'
