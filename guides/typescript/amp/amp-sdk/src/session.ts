/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Sandbox, PtyHandle } from '@daytona/sdk'
import { AmpMessage, AssistantMessage, ResultMessage, UserMessage } from './types.js'
import { renderMarkdown } from './utils.js'

const DEBUG = process.env.DEBUG === '1' || process.env.DEBUG === 'true'
function debug(...args: unknown[]) {
  if (DEBUG) console.error('[debug]', ...args)
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
  return `🔧 ${description}`
}

// Represents an Amp Code session within a Daytona sandbox
export class AmpSession {
  private threadId: string | null = null
  private systemPrompt: string | null = null
  private ptyHandle: PtyHandle | null = null
  private buffer = ''
  private onResponseComplete?: () => void

  constructor(private sandbox: Sandbox) {
    debug('AmpSession constructed. Initial threadId:', this.threadId)
  }

  // Quote a string so it is passed to a shell command literally.
  private shellQuote(s: string): string {
    return `'${s.replace(/'/g, "'\\''")}'`
  }

  // Handle a single JSON line from Amp's --stream-json output
  private handleJsonLine(line: string): void {
    try {
      const parsed = JSON.parse(line) as AmpMessage
      debug('parsed', parsed.type, parsed.subtype ?? '')

      // System message with thread ID
      if (parsed.type === 'system' && parsed.subtype === 'init') {
        const sysMsg = parsed as { session_id?: string; thread_id?: string }
        // Amp uses session_id in init message
        if (sysMsg.session_id && !this.threadId) {
          this.threadId = sysMsg.session_id
          debug('captured thread_id from init:', this.threadId)
        }
        return
      }

      // Assistant messages contain text and tool use blocks
      if (parsed.type === 'assistant') {
        const msg = parsed as AssistantMessage
        const outputs: string[] = []

        for (const block of msg.message.content) {
          if (block.type === 'text' && block.text) {
            outputs.push(renderMarkdown(block.text))
          } else if (block.type === 'tool_use') {
            outputs.push(formatToolUse(block))
          }
        }

        if (outputs.length > 0) {
          const rendered = outputs.join('\n')
          process.stdout.write(rendered.endsWith('\n') ? rendered : `${rendered}\n`)
        }
        return
      }

      // User message with tool_result = output from a tool run
      if (parsed.type === 'user') {
        const msg = parsed as UserMessage
        const blocks = msg.message?.content ?? []
        const toolResults = blocks.filter((b) => b.type === 'tool_result') as Array<{
          type: 'tool_result'
          content: string
          is_error?: boolean
        }>
        if (toolResults.length > 0) {
          const lines = toolResults.map((b) => (b.is_error ? `⚠ ${b.content}` : b.content))
          const rendered = lines.join('\n')
          process.stdout.write(rendered.endsWith('\n') ? rendered : `${rendered}\n`)
        }
        return
      }

      // Result message at end
      if (parsed.type === 'result') {
        const msg = parsed as ResultMessage
        if (msg.is_error) {
          if (msg.error?.includes('require paid credits')) {
            console.error('\n❌ Amp execute mode requires paid credits. Please add credits at https://ampcode.com/pay')
            process.exit(1)
          }
          process.stdout.write(`\n❌ Error: ${msg.error}`)
        }
        // Signal that the response is complete
        this.onResponseComplete?.()
      }
    } catch {
      // Not valid JSON, ignore
      debug('invalid JSON line:', line)
    }
  }

  // Handle streamed data from PTY
  private handleData(data: Uint8Array): void {
    // Append new data to the buffer
    this.buffer += new TextDecoder().decode(data)
    // Split the buffer into complete lines
    const lines = this.buffer.split('\n')
    // Keep any incomplete line in the buffer for next time
    this.buffer = lines.pop() || ''
    // Process each complete line
    for (const line of lines.filter((l) => l.trim())) {
      this.handleJsonLine(line)
    }
  }

  // Run an amp command via PTY and wait for completion
  private async runAmpCommand(args: string[]): Promise<void> {
    const command = ['amp', '--dangerously-allow-all', '--stream-json', '-m smart', ...args].join(' ')
    debug('running:', command)

    // Send command to the PTY
    await this.ptyHandle!.sendInput(`cd /home/daytona && ${command}\n`)

    // Wait for the response to complete (signaled by result message)
    await new Promise<void>((resolve) => {
      this.onResponseComplete = resolve
    })
  }

  // Fallback: get most recent thread ID by parsing `amp threads list` text output
  private async getThreadIdFromList(): Promise<string | null> {
    const result = await this.sandbox.process.executeCommand('amp threads list', '/home/daytona')
    if (result.exitCode !== 0 || !result.result) {
      debug('failed to list threads via text output:', result.result)
      return null
    }

    const lines = result.result
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean)
    if (lines.length <= 2) {
      return null
    }

    for (const line of lines) {
      // Skip header and separator rows
      if (line.startsWith('Title') || line.startsWith('─')) continue

      const parts = line.split(/\s{2,}/)
      const maybeId = parts[parts.length - 1]
      if (maybeId && maybeId.startsWith('T-')) {
        debug('parsed threadId from list:', maybeId)
        return maybeId
      }
    }

    return null
  }

  // Processes a user prompt by running amp CLI
  async processPrompt(prompt: string): Promise<void> {
    debug('processPrompt called. Current threadId:', this.threadId)
    console.log('Thinking...')

    if (this.threadId) {
      // Continue existing thread. Place -x before subcommand so it's treated as a global option.
      await this.runAmpCommand(['-x', this.shellQuote(prompt), 'threads', 'continue', this.threadId])
    } else {
      // Start new thread; thread/session id should be captured from streamed JSON init message
      await this.runAmpCommand(['-x', this.shellQuote(prompt)])

      // If we still don't have a threadId from the stream, fall back to parsing `amp threads list`
      if (!this.threadId) {
        this.threadId = await this.getThreadIdFromList()
      }
      debug('after initial prompt, threadId is:', this.threadId)
    }

    console.log()
  }

  // Initialize the session with an optional system prompt
  async initialize(options?: { systemPrompt?: string }): Promise<void> {
    console.log('Starting Amp Code...')

    // Create a PTY (pseudo-terminal) for streaming output from Amp
    this.ptyHandle = await this.sandbox.process.createPty({
      id: `amp-pty-${Date.now()}`,
      cols: 120,
      rows: 30,
      onData: (data: Uint8Array) => this.handleData(data),
    })
    debug('created PTY session')

    // Wait for PTY connection
    await this.ptyHandle.waitForConnection()

    if (options?.systemPrompt?.trim()) {
      this.systemPrompt = options.systemPrompt.trim()
      // Send system prompt as first message
      await this.processPrompt(`Follow these instructions for the rest of this conversation:\n\n${this.systemPrompt}`)
    }

    console.log('Agent ready. Press Ctrl+C at any time to exit.\n')
  }

  // Cleanup the PTY session
  async cleanup(): Promise<void> {
    try {
      if (this.ptyHandle) {
        await this.ptyHandle.kill()
        debug('killed PTY session')
      }
    } catch (e) {
      debug('error killing PTY session:', e)
    }
  }
}
