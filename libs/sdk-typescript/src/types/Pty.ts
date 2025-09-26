/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Options for creating a PTY session
 */
export interface PtyCreateOptions {
  /**
   * The unique identifier for the PTY session
   */
  id: string

  /**
   * Starting directory for the PTY session, defaults to the sandbox's working directory
   */
  cwd?: string

  /**
   * Environment variables for the PTY session
   */
  envs?: Record<string, string>

  /**
   * Number of terminal columns
   */
  cols?: number

  /**
   * Number of terminal rows
   */
  rows?: number
}

/**
 * Options for connecting to a PTY session
 */
export interface PtyConnectOptions {
  /**
   * Callback to handle PTY output data
   */
  onData: (data: Uint8Array) => void | Promise<void>
}

/**
 * PTY session result on exit
 */
export interface PtyResult {
  /**
   * Exit code when the PTY process ends
   */
  exitCode?: number

  /**
   * Error message if the PTY failed
   */
  error?: string
}
