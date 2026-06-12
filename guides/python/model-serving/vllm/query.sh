#!/usr/bin/env bash
# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# ENDPOINT and TOKEN are printed by serve_vllm.py.
: "${ENDPOINT:?set ENDPOINT to the preview URL}"
: "${TOKEN:?set TOKEN to the preview token}"

curl -sS --connect-timeout 30 --max-time 120 "$ENDPOINT/v1/chat/completions" \
  -H "x-daytona-preview-token: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemma-4-moe",
    "messages": [{"role": "user", "content": "Write a haiku about sandboxes for AI agents."}],
    "max_tokens": 64
  }'
