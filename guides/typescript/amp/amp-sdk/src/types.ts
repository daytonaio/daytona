/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Base message type from Amp's stream-json output
export interface AmpMessage {
  type: string
  subtype?: string
}

// System message (e.g., init)
export interface SystemMessage extends AmpMessage {
  type: 'system'
  subtype: 'init'
  session_id: string
}

// Content block types
export interface TextBlock {
  type: 'text'
  text: string
}

export interface ToolUseBlock {
  type: 'tool_use'
  id: string
  name: string
  input?: Record<string, unknown>
}

export interface ToolResultBlock {
  type: 'tool_result'
  tool_use_id: string
  content: string
  is_error?: boolean
}

export type ContentBlock = TextBlock | ToolUseBlock | ToolResultBlock

// Assistant message with content blocks
export interface AssistantMessage extends AmpMessage {
  type: 'assistant'
  message: {
    role: 'assistant'
    content: ContentBlock[]
  }
}

// User message (tool results)
export interface UserMessage extends AmpMessage {
  type: 'user'
  message: {
    role: 'user'
    content: ToolResultBlock[]
  }
}

// Result message at the end of a response
export interface ResultMessage extends AmpMessage {
  type: 'result'
  is_error?: boolean
  error?: string
  duration_ms?: number
  usage?: {
    input_tokens: number
    output_tokens: number
  }
}
