/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Event shapes emitted by `gemini --output-format stream-json` (newline-delimited JSON).
// Mirrors the Gemini CLI's documented stream-json schema:
// https://geminicli.com/docs/cli/headless/

export interface GeminiStreamEvent {
  type: string
  timestamp?: string
}

export interface InitEvent extends GeminiStreamEvent {
  type: 'init'
  session_id: string
  model: string
}

export interface MessageEvent extends GeminiStreamEvent {
  type: 'message'
  role: 'user' | 'assistant'
  content: string
  delta?: boolean
}

export interface ToolUseEvent extends GeminiStreamEvent {
  type: 'tool_use'
  tool_name: string
  tool_id: string
  parameters: Record<string, unknown>
}

export interface ToolResultEvent extends GeminiStreamEvent {
  type: 'tool_result'
  tool_id: string
  status: 'success' | 'error'
  output?: string
  error?: { type: string; message: string }
}

export interface ErrorEvent extends GeminiStreamEvent {
  type: 'error'
  severity: 'warning' | 'error'
  message: string
}

export interface ModelStreamStats {
  total_tokens: number
  input_tokens: number
  output_tokens: number
  cached: number
  input: number
}

export interface ResultEvent extends GeminiStreamEvent {
  type: 'result'
  status: 'success' | 'error'
  error?: { type: string; message: string }
  stats?: {
    total_tokens: number
    input_tokens: number
    output_tokens: number
    cached: number
    input: number
    duration_ms: number
    tool_calls: number
    models: Record<string, ModelStreamStats>
  }
}
