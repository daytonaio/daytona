/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Letta message type definitions
export interface SystemMessage {
  type: 'system'
  subtype: 'init'
}

export interface ReasoningMessage {
  type: 'message'
  message_type: 'reasoning_message'
  reasoning: string
  uuid: string
  seq_id: number
}

export interface AssistantMessage {
  type: 'message'
  message_type: 'assistant_message'
  content: string
  uuid: string
  seq_id: number
}

export interface ToolCall {
  tool_call_id: string
  name?: string
  arguments?: string
}

export interface ApprovalRequestMessage {
  type: 'message'
  message_type: 'approval_request_message'
  tool_call: ToolCall
  uuid: string
  seq_id: number
}

export interface StopReasonMessage {
  type: 'message'
  message_type: 'stop_reason'
  stop_reason: string
  uuid: string
  seq_id: number
}

export interface ResultMessage {
  type: 'result'
  result: string
  otid: string
  seq_id: number
}

export type LettaMessage =
  | SystemMessage
  | ReasoningMessage
  | AssistantMessage
  | ApprovalRequestMessage
  | StopReasonMessage
  | ResultMessage
