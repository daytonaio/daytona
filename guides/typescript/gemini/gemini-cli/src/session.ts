/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Sandbox, PtyHandle } from '@daytona/sdk'
import {
  GeminiStreamEvent,
  InitEvent,
  MessageEvent,
  ToolUseEvent,
  ToolResultEvent,
  ErrorEvent,
  ResultEvent,
} from './types.js'

const WORK_DIR = '/home/daytona'

const DEBUG = false
function debug(...args: unknown[]) {
  if (DEBUG) console.error('[debug]', ...args)
}

export class GeminiSession {
  private sessionId: string | null = null
  private ptyHandle: PtyHandle | null = null
  private buffer = ''
  // Reused across handleData calls so partial multi-byte UTF-8 sequences split
  // across PTY chunks are preserved instead of producing corrupt characters.
  private decoder = new TextDecoder('utf-8')
  private onResponseComplete?: () => void

  constructor(private sandbox: Sandbox) {}

  // Quote a string so it is passed to a shell command literally.
  private shellQuote(s: string): string {
    return `'${s.replace(/'/g, "'\\''")}'`
  }

  private handleEvent(event: GeminiStreamEvent): void {
    switch (event.type) {
      case 'init': {
        const init = event as InitEvent
        if (init.session_id && !this.sessionId) {
          this.sessionId = init.session_id
          debug('captured session_id:', this.sessionId)
        }
        return
      }
      case 'message': {
        const msg = event as MessageEvent
        if (msg.role === 'assistant' && msg.content) {
          process.stdout.write(msg.content)
        }
        return
      }
      case 'tool_use': {
        const tool = event as ToolUseEvent
        // Skip update_topic: an internal Gemini bookkeeping tool, not a user-facing action.
        if (tool.tool_name === 'update_topic') return
        process.stdout.write(`\n[tool] ${tool.tool_name}\n`)
        return
      }
      case 'tool_result': {
        const result = event as ToolResultEvent
        if (result.status === 'error' && result.error) {
          process.stdout.write(`\n[tool error] ${result.error.message}\n`)
        }
        return
      }
      case 'error': {
        const err = event as ErrorEvent
        process.stderr.write(`\n[${err.severity}] ${err.message}\n`)
        return
      }
      case 'result': {
        const res = event as ResultEvent
        if (res.status === 'error' && res.error) {
          process.stderr.write(`\nFailed: ${res.error.message}\n`)
        }
        process.stdout.write('\n')
        this.onResponseComplete?.()
        return
      }
    }
  }

  // Buffer raw PTY bytes and dispatch each complete newline-delimited JSON event.
  private handleData(data: Uint8Array): void {
    this.buffer += this.decoder.decode(data, { stream: true })
    const lines = this.buffer.split('\n')
    this.buffer = lines.pop() || ''
    for (const line of lines.map((l) => l.trim()).filter(Boolean)) {
      try {
        this.handleEvent(JSON.parse(line) as GeminiStreamEvent)
      } catch {
        debug('non-JSON line:', line)
      }
    }
  }

  async initialize(): Promise<void> {
    this.ptyHandle = await this.sandbox.process.createPty({
      id: `gemini-pty-${Date.now()}`,
      cols: 120,
      rows: 30,
      onData: (data: Uint8Array) => this.handleData(data),
    })
    await this.ptyHandle.waitForConnection()
  }

  // Run a single headless turn and resolve once Gemini emits its result event.
  async processPrompt(prompt: string): Promise<void> {
    const flags = ['-p', this.shellQuote(prompt), '--yolo', '--output-format', 'stream-json']
    // -r resumes the existing session for multi-turn continuity.
    if (this.sessionId) flags.unshift('-r', this.shellQuote(this.sessionId))
    const command = ['gemini', ...flags].join(' ')
    debug('running:', command)

    await this.ptyHandle!.sendInput(`cd ${WORK_DIR} && ${command}\n`)
    await new Promise<void>((resolve) => {
      this.onResponseComplete = resolve
    })
  }

  async cleanup(): Promise<void> {
    try {
      if (this.ptyHandle) await this.ptyHandle.kill()
    } catch (e) {
      debug('error killing PTY session:', e)
    }
  }
}
