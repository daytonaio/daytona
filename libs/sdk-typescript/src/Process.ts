/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Command,
  Configuration,
  Session,
  SessionExecuteRequest,
  SessionExecuteResponse,
  ToolboxApi,
} from '@daytonaio/api-client'
import { SandboxCodeToolbox } from './Sandbox'
import { ExecuteResponse } from './types/ExecuteResponse'
import { ArtifactParser } from './utils/ArtifactParser'
import { processStreamingResponse } from './utils/Stream'
import { Buffer } from 'buffer'

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

/**
 * Handles process and code execution within a Sandbox.
 *
 * @class
 */
export class Process {
  constructor(
    private readonly sandboxId: string,
    private readonly clientConfig: Configuration,
    private readonly codeToolbox: SandboxCodeToolbox,
    private readonly toolboxApi: ToolboxApi,
    private readonly getRootDir: () => Promise<string>,
  ) {}

  /**
   * Executes a shell command in the Sandbox.
   *
   * @param {string} command - Shell command to execute
   * @param {string} [cwd] - Working directory for command execution. If not specified, uses the Sandbox root directory.
   * Default is the user's root directory.
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

    const response = await this.toolboxApi.executeCommand(this.sandboxId, {
      command,
      timeout,
      cwd: cwd ?? (await this.getRootDir()),
    })

    // Parse artifacts from the output
    const artifacts = ArtifactParser.parseArtifacts(response.data.result)

    // Return enhanced response with parsed artifacts
    return {
      ...response.data,
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
  public async createSession(sessionId: string): Promise<void> {
    await this.toolboxApi.createSession(this.sandboxId, {
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
    const response = await this.toolboxApi.getSession(this.sandboxId, sessionId)
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
  public async getSessionCommand(sessionId: string, commandId: string): Promise<Command> {
    const response = await this.toolboxApi.getSessionCommand(this.sandboxId, sessionId, commandId)
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
   *                                           - output: Command output (if synchronous execution)
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
   * console.log(result.output);  // Prints: /home/daytona
   */
  public async executeSessionCommand(
    sessionId: string,
    req: SessionExecuteRequest,
    timeout?: number,
  ): Promise<SessionExecuteResponse> {
    const response = await this.toolboxApi.executeSessionCommand(
      this.sandboxId,
      sessionId,
      req,
      undefined,
      timeout ? { timeout: timeout * 1000 } : {},
    )
    return response.data
  }

  /**
   * Get the logs for a command executed in a session.
   *
   * @param {string} sessionId - Unique identifier of the session
   * @param {string} commandId - Unique identifier of the command
   * @returns {Promise<string>} Command logs
   *
   * @example
   * const logs = await process.getSessionCommandLogs('my-session', 'cmd-123');
   * console.log('Command output:', logs);
   */
  public async getSessionCommandLogs(sessionId: string, commandId: string): Promise<string>
  /**
   * Asynchronously retrieve and process the logs for a command executed in a session as they become available.
   *
   * @param {string} sessionId - Unique identifier of the session
   * @param {string} commandId - Unique identifier of the command
   * @param {function} onLogs - Callback function to handle each log chunk
   * @returns {Promise<void>}
   *
   * @example
   * const logs = await process.getSessionCommandLogs('my-session', 'cmd-123', (chunk) => {
   *   console.log('Log chunk:', chunk);
   * });
   */
  public async getSessionCommandLogs(
    sessionId: string,
    commandId: string,
    onLogs: (chunk: string) => void,
  ): Promise<void>
  public async getSessionCommandLogs(
    sessionId: string,
    commandId: string,
    onLogs?: (chunk: string) => void,
  ): Promise<string | void> {
    if (!onLogs) {
      const response = await this.toolboxApi.getSessionCommandLogs(this.sandboxId, sessionId, commandId)
      return response.data
    }

    const url = `${this.clientConfig.basePath}/toolbox/${this.sandboxId}/toolbox/process/session/${sessionId}/command/${commandId}/logs?follow=true`

    await processStreamingResponse(
      () => fetch(url, { method: 'GET', headers: this.clientConfig.baseOptions.headers }),
      onLogs,
      () =>
        this.getSessionCommand(sessionId, commandId).then((res) => res.exitCode !== null && res.exitCode !== undefined),
    )
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
  public async listSessions(): Promise<Session[]> {
    const response = await this.toolboxApi.listSessions(this.sandboxId)
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
  public async deleteSession(sessionId: string): Promise<void> {
    await this.toolboxApi.deleteSession(this.sandboxId, sessionId)
  }
}
