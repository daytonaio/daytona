/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Sandbox } from '@daytonaio/sdk'
import { LettaMessage, ApprovalRequestMessage, ResultMessage, ToolCall } from './types'
import { renderMarkdown } from './utils'

// Merges incoming tool call fragments into the current accumulated tool call
function accumulateToolCall(current: any, incoming: any): any {
  return current
    ? {
        ...current,
        ...incoming,
        arguments: (current.arguments || '') + (incoming.arguments || ''),
      }
    : incoming
}

// Formats a tool call for display
function formatToolCall(toolCall: any): string {
  if (!toolCall) return ''

  let description = ''

  // Generate an easy-to-read description based on tool arguments
  try {
    const args = JSON.parse(toolCall.arguments || '{}')
    description =
      args.description ||
      args.command ||
      (args.file_path && `${toolCall.name} ${args.file_path}`) ||
      (args.query && `${toolCall.name}: ${args.query}`) ||
      (args.url && `${toolCall.name} ${args.url}`) ||
      toolCall.name
  } catch (error) {
    // Fall back to a basic description and log the parse error
    description = toolCall.name || 'Tool call'
    console.warn('Failed to parse tool call arguments as JSON:', toolCall.arguments, error)
  }

  return `\n🔧 ${description}`
}

// Represents a Letta Code session within a Daytona sandbox
export class LettaSession {
  private currentToolCall: ToolCall | null = null
  private buffer = ''
  private ptyHandle: any
  private onResponseComplete?: () => void
  private onAgentInitialized?: () => void

  constructor(private sandbox: Sandbox) {}

  // Handles parsed messages from Letta's stream-json output
  handleParsedMessage(parsed: LettaMessage): string | undefined {
    // System message signals Letta has finished initializing
    if (parsed.type === 'system') {
      this.onAgentInitialized?.()
    }

    // Message types stream various parts of the agent's response
    else if (parsed.type === 'message') {
      const msgType = parsed.message_type

      // Approval request messages stream tool calls incrementally
      // Arguments arrive in multiple incomplete fragments, so we need to accumulate them
      if (msgType === 'approval_request_message') {
        const msg = parsed as ApprovalRequestMessage
        const toolCall = msg.tool_call

        // Detect when the tool call ID changes
        if (
          toolCall.tool_call_id &&
          this.currentToolCall &&
          toolCall.tool_call_id !== this.currentToolCall.tool_call_id
        ) {
          // Output the completed tool call and start accumulating the new one
          const output = formatToolCall(this.currentToolCall)
          this.currentToolCall = accumulateToolCall(null, toolCall)
          return output
        } else {
          // Accumulate tool call fragments
          this.currentToolCall = accumulateToolCall(this.currentToolCall, toolCall)
        }
      }

      // Stop reason signals that all fragments for the current tool call have been sent
      else if (msgType === 'stop_reason') {
        const output = formatToolCall(this.currentToolCall)
        this.currentToolCall = null
        return output
      }
    }

    // Result contains the agent's final formatted response after all processing
    // This is the complete output to display to the user
    else if (parsed.type === 'result') {
      const msg = parsed as ResultMessage
      this.currentToolCall = null
      this.onResponseComplete?.()
      return `\n${renderMarkdown(msg.result)}`
    }
  }

  // Handle streamed JSON data from Letta Code
  handleData(data: Uint8Array): void {
    // Append new data to the buffer
    this.buffer += new TextDecoder().decode(data)
    // Split the buffer into complete lines
    const lines = this.buffer.split('\n')
    // Keep any incomplete line in the buffer for next time
    this.buffer = lines.pop() || ''
    // Process each complete line
    for (const line of lines.filter((l) => l.trim())) {
      try {
        const output = this.handleParsedMessage(JSON.parse(line))
        if (output) process.stdout.write(output)
      } catch {}
    }
  }

  // Processes a user prompt by sending it to Letta and waiting for a response
  async processPrompt(prompt: string): Promise<void> {
    console.log('Thinking...')

    // Send the user's message to Letta in stream-json format
    await this.ptyHandle.sendInput(
      JSON.stringify({
        type: 'user',
        message: { role: 'user', content: prompt },
      }) + '\n',
    )

    // Wait for the response to complete
    await new Promise<void>((resolve) => {
      this.onResponseComplete = resolve
    })

    console.log('\n')
  }

  // Initializes the Letta Code session
  async initialize(systemPrompt: string): Promise<void> {
    console.log('Starting Letta Code...')

    // Create a PTY (pseudo-terminal) for bidirectional communication with Letta
    this.ptyHandle = await this.sandbox.process.createPty({
      id: `letta-pty-${Date.now()}`,
      cols: 120,
      rows: 30,
      onData: (data: Uint8Array) => this.handleData(data),
    })

    // Wait for PTY connection
    await this.ptyHandle.waitForConnection()

    // Start Letta Code command in the PTY with custom system prompt
    await this.ptyHandle.sendInput(
      `letta --new --system-custom "${systemPrompt.replace(/"/g, '\\"')}" --input-format stream-json --output-format stream-json --yolo -p\n`,
    )

    // Wait for agent to initialize
    console.log('Initializing agent...')
    await new Promise<void>((resolve) => {
      this.onAgentInitialized = resolve
    })
    console.log('Agent initialized. Press Ctrl+C at any time to exit.\n')
  }
}
