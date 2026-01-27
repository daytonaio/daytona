/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Configuration,
  ProcessApi,
  Command,
  Session,
  SessionExecuteRequest,
  SessionExecuteResponse as ApiSessionExecuteResponse,
  PtyCreateRequest,
  PtySessionInfo,
} from '@daytonaio/toolbox-api-client'
import { SandboxCodeToolbox } from './Sandbox'
import { ExecuteResponse } from './types/ExecuteResponse'
import { ArtifactParser } from './utils/ArtifactParser'
import { stdDemuxStream } from './utils/Stream'
import { Buffer } from 'buffer'
import { PtyHandle } from './PtyHandle'
import { PtyCreateOptions, PtyConnectOptions } from './types/Pty'
import { createSandboxWebSocket } from './utils/WebSocket'
import { WithInstrumentation } from './utils/otel.decorator'

// 3-byte multiplexing markers inserted by the shell labelers
export const STDOUT_PREFIX_BYTES = new Uint8Array([0x01, 0x01, 0x01])
export const STDERR_PREFIX_BYTES = new Uint8Array([0x02, 0x02, 0x02])
export const MAX_PREFIX_LEN = Math.max(STDOUT_PREFIX_BYTES.length, STDERR_PREFIX_BYTES.length)

/**
 * Parameters for code execution.
 */
export class CodeRunParams {
  /**
   * Command line arguments
   */
  argv?: string[]
  /**
   * Environment variables
   */
  env?: Record<string, string>
}

export interface SessionExecuteResponse extends ApiSessionExecuteResponse {
  stdout?: string
  stderr?: string
}

export interface SessionCommandLogsResponse {
  output?: string
  stdout?: string
  stderr?: string
}

/**
 * Handles process and code execution within a Sandbox.
 *
 * @class
 */
export class Process {
  constructor(
    private readonly clientConfig: Configuration,
    private readonly codeToolbox: SandboxCodeToolbox,
    private readonly apiClient: ProcessApi,
    private readonly getPreviewToken: () => Promise<string>,
    private readonly ensureToolboxUrl: () => Promise<void>,
  ) {}

  /**
   * Executes a shell command in the Sandbox.
   *
   * @param {string} command - Shell command to execute
   * @param {string} [cwd] - Working directory for command execution. If not specified, uses the sandbox working directory.
   * @param {Record<string, string>} [env] - Environment variables to set for the command
   * @param {number} [timeout] - Maximum time in seconds to wait for the command to complete. 0 means wait indefinitely.
   * @returns {Promise<ExecuteResponse>} Command execution results containing:
   *                                    - exitCode: The command's exit status
   *                                    - result: Standard output from the command
   *                                    - artifacts: ExecutionArtifacts object containing `stdout` (same as result) and `charts` (matplotlib charts metadata)
   *
   * @example
   * // Simple command
   * const response = await process.executeCommand('echo "Hello"');
   * console.log(response.artifacts.stdout);  // Prints: Hello
   *
   * @example
   * // Command with working directory
   * const result = await process.executeCommand('ls', 'workspace/src');
   *
   * @example
   * // Command with timeout
   * const result = await process.executeCommand('sleep 10', undefined, 5);
   */
  @WithInstrumentation()
  public async executeCommand(
    command: string,
    cwd?: string,
    env?: Record<string, string>,
    timeout?: number,
  ): Promise<ExecuteResponse> {
    const base64UserCmd = Buffer.from(command).toString('base64')
    command = `echo '${base64UserCmd}' | base64 -d | sh`

    if (env && Object.keys(env).length > 0) {
      const safeEnvExports =
        Object.entries(env)
          .map(([key, value]) => {
            const encodedValue = Buffer.from(value).toString('base64')
            return `export ${key}=$(echo '${encodedValue}' | base64 -d)`
          })
          .join(';') + ';'
      command = `${safeEnvExports} ${command}`
    }

    command = `sh -c "${command}"`

    const response = await this.apiClient.executeCommand({
      command,
      timeout,
      cwd: cwd,
    })

    // Parse artifacts from the output
    const artifacts = ArtifactParser.parseArtifacts(response.data.result)

    // Return enhanced response with parsed artifacts
    return {
      exitCode: response.data.exitCode ?? (response.data as any).code,
      result: artifacts.stdout,
      artifacts,
    }
  }

  /**
   * Executes code in the Sandbox using the appropriate language runtime.
   *
   * @param {string} code - Code to execute
   * @param {CodeRunParams} params - Parameters for code execution
   * @param {number} [timeout] - Maximum time in seconds to wait for execution to complete
   * @returns {Promise<ExecuteResponse>} Code execution results containing:
   *                                    - exitCode: The execution's exit status
   *                                    - result: Standard output from the code
   *                                    - artifacts: ExecutionArtifacts object containing `stdout` (same as result) and `charts` (matplotlib charts metadata)
   *
   * @example
   * // Run TypeScript code
   * const response = await process.codeRun(`
   *   const x = 10;
   *   const y = 20;
   *   console.log(\`Sum: \${x + y}\`);
   * `);
   * console.log(response.artifacts.stdout);  // Prints: Sum: 30
   *
   * @example
   * // Run Python code with matplotlib
   * const response = await process.codeRun(`
   * import matplotlib.pyplot as plt
   * import numpy as np
   *
   * x = np.linspace(0, 10, 30)
   * y = np.sin(x)
   *
   * plt.figure(figsize=(8, 5))
   * plt.plot(x, y, 'b-', linewidth=2)
   * plt.title('Line Chart')
   * plt.xlabel('X-axis (seconds)')
   * plt.ylabel('Y-axis (amplitude)')
   * plt.grid(True)
   * plt.show()
   * `);
   *
   * if (response.artifacts?.charts) {
   *   const chart = response.artifacts.charts[0];
   *
   *   console.log(`Type: ${chart.type}`);
   *   console.log(`Title: ${chart.title}`);
   *   if (chart.type === ChartType.LINE) {
   *     const lineChart = chart as LineChart
   *     console.log('X Label:', lineChart.x_label)
   *     console.log('Y Label:', lineChart.y_label)
   *     console.log('X Ticks:', lineChart.x_ticks)
   *     console.log('Y Ticks:', lineChart.y_ticks)
   *     console.log('X Tick Labels:', lineChart.x_tick_labels)
   *     console.log('Y Tick Labels:', lineChart.y_tick_labels)
   *     console.log('X Scale:', lineChart.x_scale)
   *     console.log('Y Scale:', lineChart.y_scale)
   *     console.log('Elements:')
   *     console.dir(lineChart.elements, { depth: null })
   *   }
   * }
   */
  @WithInstrumentation()
  public async codeRun(code: string, params?: CodeRunParams, timeout?: number): Promise<ExecuteResponse> {
    const runCommand = this.codeToolbox.getRunCommand(code, params)
    return this.executeCommand(runCommand, undefined, params?.env, timeout)
  }

  /**
   * Creates a new long-running background session in the Sandbox.
   *
   * Sessions are background processes that maintain state between commands, making them ideal for
   * scenarios requiring multiple related commands or persistent environment setup. You can run
   * long-running commands and monitor process status.
   *
   * @param {string} sessionId - Unique identifier for the new session
   * @returns {Promise<void>}
   *
   * @example
   * // Create a new session
   * const sessionId = 'my-session';
   * await process.createSession(sessionId);
   * const session = await process.getSession(sessionId);
   * // Do work...
   * await process.deleteSession(sessionId);
   */
  @WithInstrumentation()
  public async createSession(sessionId: string): Promise<void> {
    await this.apiClient.createSession({
      sessionId,
    })
  }

  /**
   * Get a session in the sandbox.
   *
   * @param {string} sessionId - Unique identifier of the session to retrieve
   * @returns {Promise<Session>} Session information including:
   *                            - sessionId: The session's unique identifier
   *                            - commands: List of commands executed in the session
   *
   * @example
   * const session = await process.getSession('my-session');
   * session.commands.forEach(cmd => {
   *   console.log(`Command: ${cmd.command}`);
   * });
   */
  public async getSession(sessionId: string): Promise<Session> {
    const response = await this.apiClient.getSession(sessionId)
    return response.data
  }

  /**
   * Gets information about a specific command executed in a session.
   *
   * @param {string} sessionId - Unique identifier of the session
   * @param {string} commandId - Unique identifier of the command
   * @returns {Promise<Command>} Command information including:
   *                            - id: The command's unique identifier
   *                            - command: The executed command string
   *                            - exitCode: Command's exit status (if completed)
   *
   * @example
   * const cmd = await process.getSessionCommand('my-session', 'cmd-123');
   * if (cmd.exitCode === 0) {
   *   console.log(`Command ${cmd.command} completed successfully`);
   * }
   */
  @WithInstrumentation()
  public async getSessionCommand(sessionId: string, commandId: string): Promise<Command> {
    const response = await this.apiClient.getSessionCommand(sessionId, commandId)
    return response.data
  }

  /**
   * Executes a command in an existing session.
   *
   * @param {string} sessionId - Unique identifier of the session to use
   * @param {SessionExecuteRequest} req - Command execution request containing:
   *                                     - command: The command to execute
   *                                     - runAsync: Whether to execute asynchronously
   * @param {number} timeout - Timeout in seconds
   * @returns {Promise<SessionExecuteResponse>} Command execution results containing:
   *                                           - cmdId: Unique identifier for the executed command
   *                                           - output: Combined command output (stdout and stderr) (if synchronous execution)
   *                                           - stdout: Standard output from the command
   *                                           - stderr: Standard error from the command
   *                                           - exitCode: Command exit status (if synchronous execution)
   *
   * @example
   * // Execute commands in sequence, maintaining state
   * const sessionId = 'my-session';
   *
   * // Change directory
   * await process.executeSessionCommand(sessionId, {
   *   command: 'cd /home/daytona'
   * });
   *
   * // Run command in new directory
   * const result = await process.executeSessionCommand(sessionId, {
   *   command: 'pwd'
   * });
   * console.log('[STDOUT]:', result.stdout);
   * console.log('[STDERR]:', result.stderr);
   */
  @WithInstrumentation()
  public async executeSessionCommand(
    sessionId: string,
    req: SessionExecuteRequest,
    timeout?: number,
  ): Promise<SessionExecuteResponse> {
    const response = await this.apiClient.sessionExecuteCommand(
      sessionId,
      req,
      timeout ? { timeout: timeout * 1000 } : {},
    )

    // Demux the output if it exists
    if (response.data.output) {
      // Convert string to bytes for demuxing
      const outputBytes = new TextEncoder().encode(response.data.output)
      const demuxedCommandLogs = parseSessionCommandLogs(outputBytes)
      return {
        ...response.data,
        stdout: demuxedCommandLogs.stdout,
        stderr: demuxedCommandLogs.stderr,
      }
    }

    return response.data
  }

  /**
   * Get the logs for a command executed in a session.
   *
   * @param {string} sessionId - Unique identifier of the session
   * @param {string} commandId - Unique identifier of the command
   * @returns {Promise<SessionCommandLogsResponse>} Command logs containing: output (combined stdout and stderr), stdout and stderr
   *
   * @example
   * const logs = await process.getSessionCommandLogs('my-session', 'cmd-123');
   * console.log('[STDOUT]:', logs.stdout);
   * console.log('[STDERR]:', logs.stderr);
   */
  public async getSessionCommandLogs(sessionId: string, commandId: string): Promise<SessionCommandLogsResponse>
  /**
   * Asynchronously retrieve and process the logs for a command executed in a session as they become available.
   *
   * @param {string} sessionId - Unique identifier of the session
   * @param {string} commandId - Unique identifier of the command
   * @param {function} onStdout - Callback function to handle stdout log chunks
   * @param {function} onStderr - Callback function to handle stderr log chunks
   * @returns {Promise<void>}
   *
   * @example
   * const logs = await process.getSessionCommandLogs('my-session', 'cmd-123', (chunk) => {
   *   console.log('[STDOUT]:', chunk);
   * }, (chunk) => {
   *   console.log('[STDERR]:', chunk);
   * });
   */
  public async getSessionCommandLogs(
    sessionId: string,
    commandId: string,
    onStdout: (chunk: string) => void,
    onStderr: (chunk: string) => void,
  ): Promise<void>

  @WithInstrumentation()
  public async getSessionCommandLogs(
    sessionId: string,
    commandId: string,
    onStdout?: (chunk: string) => void,
    onStderr?: (chunk: string) => void,
  ): Promise<SessionCommandLogsResponse | void> {
    if (!onStdout && !onStderr) {
      const response = await this.apiClient.getSessionCommandLogs(sessionId, commandId)

      // Parse the response data if it's available
      if (response.data) {
        // Convert string to bytes for demuxing
        const outputBytes = new TextEncoder().encode(response.data || '')
        const demuxedCommandLogs = parseSessionCommandLogs(outputBytes)

        return {
          output: response.data,
          stdout: demuxedCommandLogs.stdout,
          stderr: demuxedCommandLogs.stderr,
        }
      }

      return {
        output: response.data,
      }
    }

    await this.ensureToolboxUrl()
    const url = `${this.clientConfig.basePath.replace(/^http/, 'ws')}/process/session/${sessionId}/command/${commandId}/logs?follow=true`

    const ws = await createSandboxWebSocket(url, this.clientConfig.baseOptions?.headers || {}, this.getPreviewToken)

    await stdDemuxStream(ws, onStdout, onStderr)
  }

  /**
   * Lists all active sessions in the Sandbox.
   *
   * @returns {Promise<Session[]>} Array of active sessions
   *
   * @example
   * const sessions = await process.listSessions();
   * sessions.forEach(session => {
   *   console.log(`Session ${session.sessionId}:`);
   *   session.commands.forEach(cmd => {
   *     console.log(`- ${cmd.command} (${cmd.exitCode})`);
   *   });
   * });
   */
  @WithInstrumentation()
  public async listSessions(): Promise<Session[]> {
    const response = await this.apiClient.listSessions()
    return response.data
  }

  /**
   * Delete a session from the Sandbox.
   *
   * @param {string} sessionId - Unique identifier of the session to delete
   * @returns {Promise<void>}
   *
   * @example
   * // Clean up a completed session
   * await process.deleteSession('my-session');
   */
  @WithInstrumentation()
  public async deleteSession(sessionId: string): Promise<void> {
    await this.apiClient.deleteSession(sessionId)
  }

  /**
   * Create a new PTY (pseudo-terminal) session in the sandbox.
   *
   * Creates an interactive terminal session that can execute commands and handle user input.
   * The PTY session behaves like a real terminal, supporting features like command history.
   *
   * @param {PtyCreateOptions & PtyConnectOptions} options - PTY session configuration including creation and connection options
   * @returns {Promise<PtyHandle>} PTY handle for managing the session
   *
   * @example
   * // Create a PTY session with custom configuration
   * const ptyHandle = await process.createPty({
   *   id: 'my-interactive-session',
   *   cwd: '/workspace',
   *   envs: { TERM: 'xterm-256color', LANG: 'en_US.UTF-8' },
   *   cols: 120,
   *   rows: 30,
   *   onData: (data) => {
   *     // Handle terminal output
   *     const text = new TextDecoder().decode(data);
   *     process.stdout.write(text);
   *   },
   * });
   *
   * // Wait for connection to be established
   * await ptyHandle.waitForConnection();
   *
   * // Send commands to the terminal
   * await ptyHandle.sendInput('ls -la\n');
   * await ptyHandle.sendInput('echo "Hello, PTY!"\n');
   * await ptyHandle.sendInput('exit\n');
   *
   * // Wait for completion and get result
   * const result = await ptyHandle.wait();
   * console.log(`PTY session completed with exit code: ${result.exitCode}`);
   *
   * // Clean up
   * await ptyHandle.disconnect();
   */
  @WithInstrumentation()
  public async createPty(options?: PtyCreateOptions & PtyConnectOptions): Promise<PtyHandle> {
    const request: PtyCreateRequest = {
      id: options.id,
      cwd: options.cwd,
      envs: options.envs,
      cols: options.cols,
      rows: options.rows,
      lazyStart: true,
    }

    const response = await this.apiClient.createPtySession(request)

    return await this.connectPty(response.data.sessionId, options)
  }

  /**
   * Connect to an existing PTY session in the sandbox.
   *
   * Establishes a WebSocket connection to an existing PTY session, allowing you to
   * interact with a previously created terminal session.
   *
   * @param {string} sessionId - ID of the PTY session to connect to
   * @param {PtyConnectOptions} options - Options for the connection including data handler
   * @returns {Promise<PtyHandle>} PTY handle for managing the session
   *
   * @example
   * // Connect to an existing PTY session
   * const handle = await process.connectPty('my-session', {
   *   onData: (data) => {
   *     // Handle terminal output
   *     const text = new TextDecoder().decode(data);
   *     process.stdout.write(text);
   *   },
   * });
   *
   * // Wait for connection to be established
   * await handle.waitForConnection();
   *
   * // Send commands to the existing session
   * await handle.sendInput('pwd\n');
   * await handle.sendInput('ls -la\n');
   * await handle.sendInput('exit\n');
   *
   * // Wait for completion
   * const result = await handle.wait();
   * console.log(`Session exited with code: ${result.exitCode}`);
   *
   * // Clean up
   * await handle.disconnect();
   */
  @WithInstrumentation()
  public async connectPty(sessionId: string, options?: PtyConnectOptions): Promise<PtyHandle> {
    // Get preview link for WebSocket connection
    await this.ensureToolboxUrl()
    const url = `${this.clientConfig.basePath.replace(/^http/, 'ws')}/process/pty/${sessionId}/connect`

    const ws = await createSandboxWebSocket(url, this.clientConfig.baseOptions?.headers || {}, this.getPreviewToken)

    const handle = new PtyHandle(
      ws,
      (cols: number, rows: number) => this.resizePtySession(sessionId, cols, rows),
      () => this.killPtySession(sessionId),
      options.onData,
      sessionId,
    )
    await handle.waitForConnection()
    return handle
  }

  /**
   * List all PTY sessions in the sandbox.
   *
   * Retrieves information about all PTY sessions, both active and inactive,
   * that have been created in this sandbox.
   *
   * @returns {Promise<PtySessionInfo[]>} Array of PTY session information
   *
   * @example
   * // List all PTY sessions
   * const sessions = await process.listPtySessions();
   *
   * for (const session of sessions) {
   *   console.log(`Session ID: ${session.id}`);
   *   console.log(`Active: ${session.active}`);
   *   console.log(`Created: ${session.createdAt}`);
   *   }
   *   console.log('---');
   * }
   */
  @WithInstrumentation()
  public async listPtySessions(): Promise<PtySessionInfo[]> {
    return (await this.apiClient.listPtySessions()).data.sessions
  }

  /**
   * Get detailed information about a specific PTY session.
   *
   * Retrieves comprehensive information about a PTY session including its current state,
   * configuration, and metadata.
   *
   * @param {string} sessionId - ID of the PTY session to retrieve information for
   * @returns {Promise<PtySessionInfo>} PTY session information
   *
   * @throws {Error} If the PTY session doesn't exist
   *
   * @example
   * // Get details about a specific PTY session
   * const session = await process.getPtySessionInfo('my-session');
   *
   * console.log(`Session ID: ${session.id}`);
   * console.log(`Active: ${session.active}`);
   * console.log(`Working Directory: ${session.cwd}`);
   * console.log(`Terminal Size: ${session.cols}x${session.rows}`);
   *
   * if (session.processId) {
   *   console.log(`Process ID: ${session.processId}`);
   * }
   */
  @WithInstrumentation()
  public async getPtySessionInfo(sessionId: string): Promise<PtySessionInfo> {
    return (await this.apiClient.getPtySession(sessionId)).data
  }

  /**
   * Kill a PTY session and terminate its associated process.
   *
   * Forcefully terminates the PTY session and cleans up all associated resources.
   * This will close any active connections and kill the underlying shell process.
   *
   * @param {string} sessionId - ID of the PTY session to kill
   * @returns {Promise<void>}
   *
   * @throws {Error} If the PTY session doesn't exist or cannot be killed
   *
   * @note This operation is irreversible. Any unsaved work in the terminal session will be lost.
   *
   * @example
   * // Kill a specific PTY session
   * await process.killPtySession('my-session');
   *
   * // Verify the session is no longer active
   * try {
   *   const info = await process.getPtySessionInfo('my-session');
   *   console.log(`Session still exists but active: ${info.active}`);
   * } catch (error) {
   *   console.log('Session has been completely removed');
   * }
   */
  @WithInstrumentation()
  public async killPtySession(sessionId: string): Promise<void> {
    await this.apiClient.deletePtySession(sessionId)
  }

  /**
   * Resize a PTY session's terminal dimensions.
   *
   * Changes the terminal size of an active PTY session. This is useful when the
   * client terminal is resized or when you need to adjust the display for different
   * output requirements.
   *
   * @param {string} sessionId - ID of the PTY session to resize
   * @param {number} cols - New number of terminal columns
   * @param {number} rows - New number of terminal rows
   * @returns {Promise<PtySessionInfo>} Updated session information reflecting the new terminal size
   *
   * @throws {Error} If the PTY session doesn't exist or resize operation fails
   *
   * @note The resize operation will send a SIGWINCH signal to the shell process,
   * allowing terminal applications to adapt to the new size.
   *
   * @example
   * // Resize a PTY session to a larger terminal
   * const updatedInfo = await process.resizePtySession('my-session', 150, 40);
   * console.log(`Terminal resized to ${updatedInfo.cols}x${updatedInfo.rows}`);
   *
   * // You can also use the PtyHandle's resize method
   * await ptyHandle.resize(150, 40); // cols, rows
   */
  @WithInstrumentation()
  public async resizePtySession(sessionId: string, cols: number, rows: number): Promise<PtySessionInfo> {
    return (await this.apiClient.resizePtySession(sessionId, { cols, rows })).data
  }
}

/**
 * Parse combined stdout/stderr output into separate streams.
 *
 * @param data - Combined log bytes with STDOUT_PREFIX_BYTES and STDERR_PREFIX_BYTES markers
 * @returns Object with separated stdout and stderr strings
 */
function parseSessionCommandLogs(data: Uint8Array): SessionCommandLogsResponse {
  const [stdoutBytes, stderrBytes] = demuxLog(data)

  // Convert bytes to strings, ignoring potential encoding issues
  const stdoutStr = new TextDecoder('utf-8', { fatal: false }).decode(stdoutBytes)
  const stderrStr = new TextDecoder('utf-8', { fatal: false }).decode(stderrBytes)

  // For backwards compatibility, output field contains the original combined data
  const outputStr = new TextDecoder('utf-8', { fatal: false }).decode(data)

  return {
    output: outputStr,
    stdout: stdoutStr,
    stderr: stderrStr,
  }
}

/**
 * Demultiplex combined stdout/stderr log data.
 *
 * @param data - Combined log bytes with STDOUT_PREFIX_BYTES and STDERR_PREFIX_BYTES markers
 * @returns Tuple of [stdout_bytes, stderr_bytes]
 */
function demuxLog(data: Uint8Array): [Uint8Array, Uint8Array] {
  const outChunks: Uint8Array[] = []
  const errChunks: Uint8Array[] = []
  let state: 'none' | 'stdout' | 'stderr' = 'none'

  // Forward index (no per-loop re-slicing)
  let i = 0

  while (i < data.length) {
    // Find the nearest forward marker (stdout or stderr) from current index
    const stdoutIndex = findSubarray(data, STDOUT_PREFIX_BYTES, i)
    const stderrIndex = findSubarray(data, STDERR_PREFIX_BYTES, i)

    // Pick the closest marker index and type
    let nextIdx = -1
    let nextMarker: 'stdout' | 'stderr' | null = null
    let nextLen = 0

    if (stdoutIndex !== -1 && (stderrIndex === -1 || stdoutIndex < stderrIndex)) {
      nextIdx = stdoutIndex
      nextMarker = 'stdout'
      nextLen = STDOUT_PREFIX_BYTES.length
    } else if (stderrIndex !== -1) {
      nextIdx = stderrIndex
      nextMarker = 'stderr'
      nextLen = STDERR_PREFIX_BYTES.length
    }

    if (nextIdx === -1) {
      // No more markers → dump remainder into current state
      if (state === 'stdout') {
        outChunks.push(data.subarray(i))
      } else if (state === 'stderr') {
        errChunks.push(data.subarray(i))
      }
      break
    }

    // Write everything before the marker into current state
    if (state === 'stdout' && nextIdx > i) {
      outChunks.push(data.subarray(i, nextIdx))
    } else if (state === 'stderr' && nextIdx > i) {
      errChunks.push(data.subarray(i, nextIdx))
    }

    // Advance past marker and switch state
    i = nextIdx + nextLen
    if (nextMarker) {
      state = nextMarker
    }
  }

  // Concatenate all chunks
  return [concatenateUint8Arrays(outChunks), concatenateUint8Arrays(errChunks)]
}

/**
 * Efficiently concatenate multiple Uint8Array chunks into a single Uint8Array.
 *
 * @param chunks - Array of Uint8Array chunks to concatenate
 * @returns A single Uint8Array containing all chunks
 */
function concatenateUint8Arrays(chunks: Uint8Array[]): Uint8Array {
  if (chunks.length === 0) {
    return new Uint8Array(0)
  }

  if (chunks.length === 1) {
    return chunks[0]
  }

  const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0)

  const result = new Uint8Array(totalLength)

  let offset = 0
  for (const chunk of chunks) {
    result.set(chunk, offset)
    offset += chunk.length
  }

  return result
}

/**
 * Helper function to find a subarray within a larger array.
 *
 * @param haystack - The array to search in
 * @param needle - The subarray to find
 * @param fromIndex - starting index
 * @returns The index of the first occurrence, or -1 if not found
 */
function findSubarray(haystack: Uint8Array, needle: Uint8Array, fromIndex = 0): number {
  if (needle.length === 0) return 0
  if (haystack.length < needle.length || fromIndex < 0 || fromIndex > haystack.length - needle.length) return -1

  const limit = haystack.length - needle.length
  for (let i = fromIndex; i <= limit; i++) {
    let j = 0
    for (; j < needle.length; j++) {
      if (haystack[i + j] !== needle[j]) break
    }
    if (j === needle.length) return i
  }
  return -1
}
