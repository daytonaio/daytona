/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * @module code-interpreter
 */

import WebSocket from 'isomorphic-ws'
import { InterpreterApi, InterpreterContext } from '@daytonaio/toolbox-api-client'
import { Configuration } from '@daytonaio/api-client'
import { DaytonaError, DaytonaTimeoutError } from './errors/DaytonaError'
import { ExecutionError, ExecutionResult, RunCodeOptions } from './types/CodeInterpreter'
import { createSandboxWebSocket } from './utils/WebSocket'

type CloseEvent = {
  code: number
  reason: string
}

const WEBSOCKET_TIMEOUT_CODE = 4008

/**
 * Handles Python code interpretation and execution within a Sandbox.
 *
 * Provides methods to execute code (currently only Python) in isolated interpreter contexts,
 * manage contexts, and stream execution output via callbacks.
 *
 * For other languages, use the `codeRun` method from the `Process` interface, or execute the appropriate command directly in the sandbox terminal.
 */
export class CodeInterpreter {
  constructor(
    private readonly clientConfig: Configuration,
    private readonly apiClient: InterpreterApi,
    private readonly getPreviewToken: () => Promise<string>,
  ) {}

  /**
   * Run Python code in the sandbox.
   *
   * @param {string} code - Code to run.
   * @param {RunCodeOptions} options - Execution options (context, envs, callbacks, timeout).
   * @returns {Promise<ExecutionResult>} ExecutionResult containing stdout, stderr and optional error info.
   *
   * @example
   * ```ts
   * const handleStdout = (msg: OutputMessage) => process.stdout.write(`STDOUT: ${msg.output}`)
   * const handleStderr = (msg: OutputMessage) => process.stdout.write(`STDERR: ${msg.output}`)
   * const handleError = (err: ExecutionError) =>
   *   console.error(`ERROR: ${err.name}: ${err.value}\n${err.traceback ?? ''}`)
   *
   * const code = `
   * import sys
   * import time
   * for i in range(5):
   *     print(i)
   *     time.sleep(1)
   * sys.stderr.write("Counting done!")
   * `
   *
   * const result = await codeInterpreter.runCode(code, {
   *   onStdout: handleStdout,
   *   onStderr: handleStderr,
   *   onError: handleError,
   *   timeout: 10,
   * })
   * ```
   */
  public async runCode(code: string, options: RunCodeOptions = {}): Promise<ExecutionResult> {
    if (!code || !code.trim()) {
      throw new DaytonaError('Code is required for execution')
    }

    const url = `${this.clientConfig.basePath.replace(/^http/, 'ws')}/process/interpreter/execute`

    const ws = await createSandboxWebSocket(url, this.clientConfig.baseOptions?.headers || {}, this.getPreviewToken)

    const result: ExecutionResult = { stdout: '', stderr: '' }

    return new Promise<ExecutionResult>((resolve, reject) => {
      let settled = false

      const cleanup = () => {
        detach()
        try {
          if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
            ws.close()
          }
        } catch {
          /* ignore */
        }
      }

      const fail = (error: Error) => {
        if (settled) return
        settled = true
        cleanup()
        reject(error)
      }

      const succeed = () => {
        if (settled) return
        settled = true
        cleanup()
        resolve(result)
      }

      const handleOpen = () => {
        const payload: Record<string, unknown> = { code }

        const context = options.context
        if (context?.id) {
          payload.contextId = context.id
        }
        if (options.envs) {
          payload.envs = options.envs
        }
        payload.timeout = options.timeout

        ws.send(JSON.stringify(payload))
      }

      const handleMessage = async (event: WebSocket.MessageEvent | WebSocket.RawData | any) => {
        try {
          const text = await this.extractMessageText(event)
          if (!text) {
            return
          }

          const chunk = JSON.parse(text)
          const chunkType = chunk.type

          if (chunkType === 'stdout') {
            const stdout = chunk.text ?? ''
            result.stdout += stdout
            if (options.onStdout) {
              await options.onStdout({ output: stdout })
            }
          } else if (chunkType === 'stderr') {
            const stderr = chunk.text ?? ''
            result.stderr += stderr
            if (options.onStderr) {
              await options.onStderr({ output: stderr })
            }
          } else if (chunkType === 'error') {
            const error: ExecutionError = {
              name: chunk.name ?? '',
              value: chunk.value ?? '',
              traceback: chunk.traceback ?? '',
            }
            result.error = error
            if (options.onError) {
              await options.onError(error)
            }
          } else if (chunkType === 'control') {
            const controlText = chunk.text ?? ''
            if (controlText === 'completed' || controlText === 'interrupted') {
              succeed()
            }
          }
        } catch {
          // Ignore invalid JSON payloads
        }
      }

      const handleClose = (event: CloseEvent | number, reason?: Buffer) => {
        if (settled) return

        const { code, message } = this.normalizeCloseEvent(event, reason)

        if (code !== 1000 && code !== 1001) {
          fail(this.createCloseError(code, message))
          return
        }

        succeed()
      }

      const handleError = (error: Error) => {
        fail(new DaytonaError(`Failed to execute code: ${error.message ?? String(error)}`))
      }

      const detach = () => {
        if ('removeEventListener' in ws) {
          ws.removeEventListener('open', handleOpen as any)
          ws.removeEventListener('message', handleMessage as any)
          ws.removeEventListener('close', handleClose as any)
          ws.removeEventListener('error', handleError as any)
        }
        if ('off' in ws) {
          ;(ws as any).off('open', handleOpen)
          ;(ws as any).off('message', handleMessage)
          ;(ws as any).off('close', handleClose)
          ;(ws as any).off('error', handleError)
        } else if ('removeListener' in ws) {
          ;(ws as any).removeListener('open', handleOpen)
          ;(ws as any).removeListener('message', handleMessage)
          ;(ws as any).removeListener('close', handleClose)
          ;(ws as any).removeListener('error', handleError)
        }
      }

      if ('addEventListener' in ws) {
        ws.addEventListener('open', handleOpen as any)
        ws.addEventListener('message', handleMessage as any)
        ws.addEventListener('close', handleClose as any)
        ws.addEventListener('error', handleError as any)
      } else if ('on' in ws) {
        ;(ws as any).on('open', handleOpen)
        ;(ws as any).on('message', handleMessage)
        ;(ws as any).on('close', handleClose)
        ;(ws as any).on('error', handleError)
      } else {
        throw new DaytonaError('Unsupported WebSocket implementation')
      }
    })
  }

  /**
   * Create a new isolated interpreter context.
   *
   * @param {string} [cwd] - Working directory for the context. Uses sandbox working directory if omitted.
   *
   * @returns {Promise<InterpreterContext>} The created context.
   *
   * @example
   * ```ts
   * const ctx = await sandbox.codeInterpreter.createContext()
   * await sandbox.codeInterpreter.runCode('x = 10', { context: ctx })
   * await sandbox.codeInterpreter.deleteContext(ctx.id!)
   * ```
   */
  public async createContext(cwd?: string): Promise<InterpreterContext> {
    return (await this.apiClient.createInterpreterContext({ cwd })).data
  }

  /**
   * List all user-created interpreter contexts (default context is excluded).
   *
   * @returns {Promise<InterpreterContext[]>} List of contexts.
   *
   * @example
   * ```ts
   * const contexts = await sandbox.codeInterpreter.listContexts()
   * for (const ctx of contexts) {
   *   console.log(ctx.id, ctx.language, ctx.cwd)
   * }
   * ```
   */
  public async listContexts(): Promise<InterpreterContext[]> {
    return (await this.apiClient.listInterpreterContexts()).data.contexts ?? []
  }

  /**
   * Delete an interpreter context and shut down its worker process.
   *
   * @param {InterpreterContext} context - Context to delete.
   *
   * @example
   * ```ts
   * const ctx = await sandbox.codeInterpreter.createContext()
   * // ... use context ...
   * await sandbox.codeInterpreter.deleteContext(ctx)
   * ```
   */
  public async deleteContext(context: InterpreterContext): Promise<void> {
    await this.apiClient.deleteInterpreterContext(context.id)
  }

  private async extractMessageText(event: WebSocket.MessageEvent | WebSocket.RawData | any): Promise<string> {
    const data = event && typeof event === 'object' && 'data' in event ? event.data : event
    if (typeof data === 'string') {
      return data
    }

    if (typeof ArrayBuffer !== 'undefined' && data instanceof ArrayBuffer) {
      return new TextDecoder('utf-8').decode(new Uint8Array(data))
    }

    if (typeof Blob !== 'undefined' && data instanceof Blob) {
      return await data.text()
    }

    if (ArrayBuffer.isView(data)) {
      return new TextDecoder('utf-8').decode(new Uint8Array(data.buffer, data.byteOffset, data.byteLength))
    }

    if (data == null) {
      return ''
    }

    return data.toString()
  }

  private normalizeCloseEvent(event: CloseEvent | number, reason?: Buffer): { code: number; message: string } {
    if (typeof event === 'number') {
      return {
        code: event,
        message: reason ? reason.toString('utf-8') : '',
      }
    }

    return {
      code: event.code,
      message: event.reason,
    }
  }

  private createCloseError(code: number, message?: string): DaytonaError {
    if (code === WEBSOCKET_TIMEOUT_CODE) {
      return new DaytonaTimeoutError(
        'Execution timed out: operation exceeded the configured `timeout`. Provide a larger value if needed.',
      )
    }
    if (message) {
      return new DaytonaError(message + ` (close code ${code})`)
    }
    return new DaytonaError(`Code execution failed: WebSocket closed with code ${code}`)
  }
}
