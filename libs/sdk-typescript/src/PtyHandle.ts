/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import WebSocket from 'isomorphic-ws'
import { PtyResult } from './types/Pty'
import { DaytonaError } from './errors/DaytonaError'
import { PtySessionInfo } from '@daytonaio/toolbox-api-client'
import { WithInstrumentation } from './utils/otel.decorator'

/**
 * PTY session handle for managing a single PTY session.
 *
 * Provides methods for sending input, resizing the terminal, waiting for completion,
 * and managing the WebSocket connection to a PTY session.
 *
 * @example
 * ```typescript
 * // Create a PTY session
 * const ptyHandle = await process.createPty({
 *   id: 'my-session',
 *   cols: 120,
 *   rows: 30,
 *   onData: (data) => {
 *     const text = new TextDecoder().decode(data);
 *     process.stdout.write(text);
 *   },
 * });
 *
 *
 * // Send commands
 * await ptyHandle.sendInput('ls -la\n');
 * await ptyHandle.sendInput('exit\n');
 *
 * // Wait for completion
 * const result = await ptyHandle.wait();
 * console.log(`PTY exited with code: ${result.exitCode}`);
 *
 * // Clean up
 * await ptyHandle.disconnect();
 * ```
 */
export class PtyHandle {
  private _exitCode?: number
  private _error?: string
  private connected = false
  private connectionEstablished = false // Track control message received

  constructor(
    private readonly ws: WebSocket,
    private readonly handleResize: (cols: number, rows: number) => Promise<PtySessionInfo>,
    private readonly handleKill: () => Promise<void>,
    private readonly onPty: (data: Uint8Array) => void | Promise<void>,
    readonly sessionId: string,
  ) {
    this.setupWebSocketHandlers()
  }

  /**
   * Exit code of the PTY process (if terminated)
   */
  get exitCode(): number | undefined {
    return this._exitCode
  }

  /**
   * Error message if the PTY failed
   */
  get error(): string | undefined {
    return this._error
  }

  /**
   * Check if connected to the PTY session
   */
  isConnected(): boolean {
    return this.connected && this.ws.readyState === WebSocket.OPEN
  }

  /**
   * Wait for the WebSocket connection to be established.
   *
   * This method ensures the PTY session is ready to receive input and send output.
   * It waits for the server to confirm the connection is established.
   *
   * @throws {Error} If connection times out (10 seconds) or connection fails
   */
  @WithInstrumentation()
  async waitForConnection(): Promise<void> {
    if (this.connectionEstablished) {
      return
    }

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new DaytonaError('PTY connection timeout'))
      }, 10000) // 10 second timeout

      const checkConnection = () => {
        if (this.connectionEstablished) {
          clearTimeout(timeout)
          resolve()
        } else if (this.ws.readyState === WebSocket.CLOSED || this._error) {
          clearTimeout(timeout)
          reject(new DaytonaError(this._error || 'Connection failed'))
        } else {
          setTimeout(checkConnection, 100)
        }
      }

      checkConnection()
    })
  }

  /**
   * Send input data to the PTY session.
   *
   * Sends keyboard input or commands to the terminal session. The data will be
   * processed as if it was typed in the terminal.
   *
   * @param {string | Uint8Array} data - Input data to send (commands, keystrokes, etc.)
   * @throws {Error} If PTY is not connected or sending fails
   *
   * @example
   * // Send a command
   * await ptyHandle.sendInput('ls -la\n');
   *
   * // Send raw bytes
   * await ptyHandle.sendInput(new Uint8Array([3])); // Ctrl+C
   */
  @WithInstrumentation()
  async sendInput(data: string | Uint8Array): Promise<void> {
    if (!this.isConnected()) {
      throw new DaytonaError('PTY is not connected')
    }

    try {
      if (typeof data === 'string') {
        this.ws.send(new TextEncoder().encode(data))
      } else {
        this.ws.send(data)
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error)
      throw new DaytonaError(`Failed to send input to PTY: ${errorMessage}`)
    }
  }

  /**
   * Resize the PTY terminal dimensions.
   *
   * Changes the terminal size which will notify terminal applications
   * about the new dimensions via SIGWINCH signal.
   *
   * @param {number} cols - New number of terminal columns
   * @param {number} rows - New number of terminal rows
   *
   * @example
   * // Resize to 120x30
   * await ptyHandle.resize(120, 30);
   */
  @WithInstrumentation()
  async resize(cols: number, rows: number): Promise<PtySessionInfo> {
    return await this.handleResize(cols, rows)
  }

  /**
   * Disconnect from the PTY session and clean up resources.
   *
   * Closes the WebSocket connection and releases any associated resources.
   * Should be called when done with the PTY session.
   *
   * @example
   * // Always clean up when done
   * try {
   *   // ... use PTY session
   * } finally {
   *   await ptyHandle.disconnect();
   * }
   */
  @WithInstrumentation()
  async disconnect(): Promise<void> {
    if (this.ws) {
      try {
        this.ws.close()
      } catch {
        // Ignore close errors
      }
    }
  }

  /**
   * Wait for the PTY process to exit and return the result.
   *
   * This method blocks until the PTY process terminates and returns
   * information about how it exited.
   *
   * @returns {Promise<PtyResult>} Result containing exit code and error information
   *
   * @example
   * // Wait for process to complete
   * const result = await ptyHandle.wait();
   *
   * if (result.exitCode === 0) {
   *   console.log('Process completed successfully');
   * } else {
   *   console.log(`Process failed with code: ${result.exitCode}`);
   *   if (result.error) {
   *     console.log(`Error: ${result.error}`);
   *   }
   * }
   */
  @WithInstrumentation()
  async wait(): Promise<PtyResult> {
    return new Promise((resolve, reject) => {
      if (this._exitCode !== undefined) {
        resolve({
          exitCode: this._exitCode,
          error: this._error,
        })
        return
      }

      const checkExit = () => {
        if (this._exitCode !== undefined) {
          resolve({
            exitCode: this._exitCode,
            error: this._error,
          })
        } else if (this._error) {
          reject(new DaytonaError(this._error))
        } else {
          setTimeout(checkExit, 100)
        }
      }

      checkExit()
    })
  }

  /**
   * Kill the PTY process and terminate the session.
   *
   * Forcefully terminates the PTY session and its associated process.
   * This operation is irreversible and will cause the PTY to exit immediately.
   *
   * @throws {Error} If the kill operation fails
   *
   * @example
   * // Kill a long-running process
   * await ptyHandle.kill();
   *
   * // Wait to confirm termination
   * const result = await ptyHandle.wait();
   * console.log(`Process terminated with exit code: ${result.exitCode}`);
   */
  @WithInstrumentation()
  async kill(): Promise<void> {
    return await this.handleKill()
  }

  private setupWebSocketHandlers(): void {
    // Set binary type for binary data handling
    if ('binaryType' in this.ws) {
      this.ws.binaryType = 'arraybuffer'
    }

    // Handle WebSocket open
    const handleOpen = async () => {
      this.connected = true
    }

    // Handle WebSocket messages - control messages and PTY data
    const handleMessage = async (event: MessageEvent | any) => {
      try {
        const data = event && typeof event === 'object' && 'data' in event ? event.data : event

        if (typeof data === 'string') {
          // Try to parse as control message first
          try {
            const controlMsg = JSON.parse(data)
            if (controlMsg.type === 'control') {
              if (controlMsg.status === 'connected') {
                this.connectionEstablished = true
                return
              } else if (controlMsg.status === 'error') {
                this._error = controlMsg.error || 'Unknown connection error'
                this.connected = false
                return
              }
            }
          } catch {
            // Not a control message, treat as PTY output
          }

          // Regular PTY text output
          if (this.onPty) {
            await this.onPty(new TextEncoder().encode(data))
          }
        } else {
          // Handle binary data (terminal output)
          let bytes: Uint8Array

          if (data instanceof ArrayBuffer) {
            bytes = new Uint8Array(data)
          } else if (ArrayBuffer.isView(data)) {
            bytes = new Uint8Array(data.buffer, data.byteOffset, data.byteLength)
          } else if (data instanceof Blob) {
            const buffer = await data.arrayBuffer()
            bytes = new Uint8Array(buffer)
          } else {
            throw new DaytonaError(`Unsupported message data type: ${Object.prototype.toString.call(data)}`)
          }

          if (this.onPty) {
            await this.onPty(bytes)
          }
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error)
        throw new DaytonaError(`Error handling PTY message: ${errorMessage}`)
      }
    }

    // Handle WebSocket errors
    const handleError = async (error: any) => {
      let errorMessage: string
      if (error instanceof Error) {
        errorMessage = error.message
      } else if (error && error instanceof Event) {
        errorMessage = 'WebSocket connection error'
      } else {
        errorMessage = String(error)
      }

      this._error = errorMessage
      this.connected = false
    }

    // Handle WebSocket close - parse structured exit data
    const handleClose = async (event: CloseEvent | any) => {
      this.connected = false

      // Parse structured exit data from close reason
      if (event && event.reason) {
        try {
          const exitData = JSON.parse(event.reason)
          if (typeof exitData.exitCode === 'number') {
            this._exitCode = exitData.exitCode
            // Store exit reason if provided (undefined for exitCode 0)
            if (exitData.exitReason) {
              this._error = exitData.exitReason
            }
          }
          // Handle error messages from server (e.g., "PTY session not found")
          if (exitData.error) {
            this._error = exitData.error
          }
        } catch {
          if (event.code === 1000) {
            this._exitCode = 0
          }
        }
      }

      // Default to exit code 0 if we can't parse it and it was a normal close
      if (this._exitCode === undefined && event && event.code === 1000) {
        this._exitCode = 0
      }
    }

    // Attach event listeners based on WebSocket implementation
    if (this.ws.addEventListener) {
      // Browser WebSocket
      this.ws.addEventListener('open', handleOpen)
      this.ws.addEventListener('message', handleMessage)
      this.ws.addEventListener('error', handleError)
      this.ws.addEventListener('close', handleClose)
    } else if ('on' in this.ws && typeof this.ws.on === 'function') {
      // Node.js WebSocket
      this.ws.on('open', handleOpen)
      this.ws.on('message', handleMessage)
      this.ws.on('error', handleError)
      this.ws.on('close', handleClose)
    } else {
      throw new DaytonaError('Unsupported WebSocket implementation')
    }
  }
}
