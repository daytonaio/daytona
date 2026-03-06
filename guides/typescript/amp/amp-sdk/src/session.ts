/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Sandbox } from '@daytonaio/sdk'
import { AmpMessage, AssistantMessage, ResultMessage, UserMessage } from './types.js'
import { renderMarkdown } from './utils.js'

const DEBUG = process.env.DEBUG === '1' || process.env.DEBUG === 'true'
function debug(stream: string, ...args: unknown[]) {
  if (DEBUG) console.error(`[${stream}]`, ...args)
}

// Extract one JSON object from the start of a string (brace-matched). Returns { parsed, rest } or null.
function extractOneJson(line: string): { parsed: unknown; rest: string } | null {
  const trimmed = line.trim()
  if (!trimmed.startsWith('{')) return null
  let depth = 0
  for (let i = 0; i < trimmed.length; i++) {
    const c = trimmed[i]
    if (c === '{') depth++
    else if (c === '}') {
      depth--
      if (depth === 0) {
        try {
          const parsed = JSON.parse(trimmed.slice(0, i + 1))
          const rest = trimmed.slice(i + 1).trim()
          return { parsed, rest }
        } catch {
          return null
        }
      }
    }
  }
  return null
}

// Formats a tool call for display (Amp uses cmd, path, file_path, command, etc.)
function formatToolUse(block: { name: string; input?: Record<string, unknown> }): string {
  const inp = block.input
  const description =
    (inp?.description as string) ||
    (inp?.cmd as string) ||
    (inp?.command as string) ||
    (inp?.path && `${block.name} ${inp.path}`) ||
    (inp?.file_path && `${block.name} ${inp.file_path}`) ||
    (inp?.query && `${block.name}: ${inp.query}`) ||
    block.name
  return `\n🔧 ${description}`
}

// Represents an Amp Code session within a Daytona sandbox
export class AmpSession {
  private buffer = ''
  private ptyHandle: any
  private onResponseComplete?: (error?: string) => void
  private onAgentInitialized?: () => void
  private isSendingInput = false

  constructor(private sandbox: Sandbox) {}

  // Handles parsed messages from Amp's stream-json output
  handleParsedMessage(parsed: AmpMessage): string | undefined {
    debug('parsed', parsed.type, parsed.subtype ?? '', parsed)

    // System message signals Amp has finished initializing
    if (parsed.type === 'system' && parsed.subtype === 'init') {
      debug('stream', 'system init')
      this.onAgentInitialized?.()
      return
    }

    // Assistant messages contain text and tool use blocks
    if (parsed.type === 'assistant') {
      const msg = parsed as AssistantMessage & { message?: { stop_reason?: string } }
      const outputs: string[] = []

      const content = msg.message.content
      let lastBlockIsBackgroundCmd = false
      for (const block of content) {
        if (block.type === 'text' && block.text) {
          outputs.push(renderMarkdown(block.text))
        } else if (block.type === 'tool_use') {
          outputs.push(formatToolUse(block))
        }
      }
      // Only return control on & if it's the last action (last content block is tool_use with cmd ending in &)
      const last = content[content.length - 1]
      if (last?.type === 'tool_use') {
        const cmd = (last.input?.cmd ?? last.input?.command) as string | undefined
        lastBlockIsBackgroundCmd = typeof cmd === 'string' && cmd.trim().endsWith('&')
      }

      if (msg.message?.stop_reason === 'end_turn') {
        debug('stream', 'assistant stop_reason end_turn')
        this.onResponseComplete?.()
      } else if (lastBlockIsBackgroundCmd) {
        // Background command as last action; return control now instead of waiting for tool result / end_turn
        debug('stream', 'assistant background cmd (&) as last action, returning control')
        this.onResponseComplete?.()
      }
      if (outputs.length > 0) {
        const out = outputs.join('')
        debug('stream', 'assistant output', out.length, 'chars')
        return out
      }
    }

    // User message with tool_result = output from a tool run. Show it (do not complete here—we only complete on end_turn, result, or & as last action).
    else if (parsed.type === 'user') {
      const msg = parsed as UserMessage
      const blocks = msg.message?.content ?? []
      const toolResults = blocks.filter((b) => b.type === 'tool_result') as Array<{
        type: 'tool_result'
        content: string
        is_error?: boolean
      }>
      if (toolResults.length > 0) {
        const lines = toolResults.map((b) => (b.is_error ? `\n⚠ ${b.content}` : `\n${b.content}`))
        debug('stream', 'tool result', toolResults.length, 'blocks')
        return lines.join('')
      }
    }

    // Result contains the final status after all processing
    else if (parsed.type === 'result') {
      const msg = parsed as ResultMessage
      debug('stream', 'result', { is_error: msg.is_error, error: msg.error, duration_ms: msg.duration_ms })

      if (msg.is_error) {
        // Check for paid credits error
        if (msg.error?.includes('require paid credits')) {
          this.onResponseComplete?.(
            'Amp execute mode requires paid credits. Please add credits at https://ampcode.com/pay',
          )
          return
        }
        this.onResponseComplete?.()
        return `\n❌ Error: ${msg.error}`
      }

      this.onResponseComplete?.()
      return ''
    }
  }

  // Handle streamed JSON data from Amp CLI
  handleData(data: Uint8Array): void {
    const text = new TextDecoder().decode(data)
    debug('pty', 'raw', data.length, 'bytes', text)
    this.buffer += text

    // Split the buffer into complete lines
    const lines = this.buffer.split('\n')
    // Keep any incomplete line in the buffer for next time
    this.buffer = lines.pop() || ''
    // Process each line; a line may contain multiple JSON objects (e.g. result + user with no newline)
    for (const line of lines) {
      let rest = line.trim()
      while (rest && rest.startsWith('{')) {
        const one = extractOneJson(rest)
        if (!one) break
        const parsed = one.parsed as AmpMessage
        rest = one.rest
        if (this.isSendingInput) {
          if (parsed.type === 'user') {
            debug('stream', 'skipped user echo')
            this.buffer = ''
            continue
          }
          this.isSendingInput = false
        }
        const output = this.handleParsedMessage(parsed)
        if (output) {
          debug('stdout', output.length, 'chars', output)
          process.stdout.write(output)
        }
      }
    }
  }

  // Processes a user prompt by sending it to the running Amp session
  async processPrompt(prompt: string): Promise<void> {
    console.log('Thinking...')

    // Send the user's message in stream-json-input format (see ampcode.com/manual#cli-streaming-json)
    const message = JSON.stringify({
      type: 'user',
      message: {
        role: 'user',
        content: [{ type: 'text', text: prompt }],
      },
    })

    this.isSendingInput = true
    await this.ptyHandle.sendInput(message + '\n')

    // Wait for end_turn or result; we can't know from tool_use alone if the agent will send more. Timeout so we don't hang if Amp never sends.
    const completionTimeoutMs = 120_000
    const error = await new Promise<string | undefined>((resolve) => {
      let settled = false
      const done = (err?: string) => {
        if (settled) return
        settled = true
        clearTimeout(timer)
        this.onResponseComplete = undefined
        resolve(err)
      }
      this.onResponseComplete = done
      const timer = setTimeout(() => {
        debug('stream', 'completion timeout after', completionTimeoutMs, 'ms')
        done()
      }, completionTimeoutMs)
    })

    if (error) {
      console.error(`\n❌ ${error}`)
      process.exit(1)
    }

    console.log('\n')
  }

  // Optional system prompt sent as the first user message so the agent has context (Amp has no --system-prompt; AGENTS.md is the other option).
  async initialize(options?: { systemPrompt?: string }): Promise<void> {
    console.log('Starting Amp Code...')

    // Create a PTY (pseudo-terminal) for bidirectional communication with Amp
    this.ptyHandle = await this.sandbox.process.createPty({
      id: 'amp-pty',
      cols: 200,
      rows: 50,
      onData: (data: Uint8Array) => this.handleData(data),
    })

    // Wait for PTY connection
    await this.ptyHandle.waitForConnection()

    // Wait a moment for the shell to be ready
    await new Promise((resolve) => setTimeout(resolve, 500))

    // Start Amp CLI: --stream-json requires --execute; with --stream-json-input,
    // messages come from stdin and Amp exits only when stdin is closed.
    const ampCommand = [
      'amp',
      '--dangerously-allow-all',
      '--execute',
      '--stream-json',
      '--stream-json-input',
      '-m smart',
    ].join(' ')

    console.log('Initializing agent...')
    await this.ptyHandle.sendInput(`${ampCommand}\n`)

    // Wait for the init system message from Amp
    await new Promise<void>((resolve) => {
      this.onAgentInitialized = resolve
    })

    if (options?.systemPrompt?.trim()) {
      await this.processPrompt(
        `Follow these instructions for the rest of this conversation:\n\n${options.systemPrompt.trim()}`,
      )
    }

    console.log('Agent ready. Press Ctrl+C at any time to exit.\n')
  }
}
