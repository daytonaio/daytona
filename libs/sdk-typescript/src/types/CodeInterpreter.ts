/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * @module code-interpreter
 */

import { InterpreterContext } from '@daytonaio/toolbox-api-client'

/**
 * Represents stdout or stderr output from code execution.
 */
export interface OutputMessage {
  /**
   * Output content.
   */
  output: string
}

/**
 * Represents an error that occurred during code execution.
 */
export interface ExecutionError {
  /**
   * Error type/class name (e.g., "ValueError", "SyntaxError").
   */
  name: string
  /**
   * Error value/message.
   */
  value: string
  /**
   * Full traceback for the error, if available.
   */
  traceback?: string
}

/**
 * Result of code execution.
 */
export interface ExecutionResult {
  /**
   * Standard output captured during execution.
   */
  stdout: string
  /**
   * Standard error captured during execution.
   */
  stderr: string
  /**
   * Details about an execution error, if one occurred.
   */
  error?: ExecutionError
}

/**
 * Options for executing code in the interpreter.
 */
export interface RunCodeOptions {
  /**
   * Interpreter context to run code in.
   */
  context?: InterpreterContext
  /**
   * Environment variables for this execution.
   */
  envs?: Record<string, string>
  /**
   * Timeout in seconds. Set to 0 for no timeout. Default is 10 minutes.
   */
  timeout?: number
  /**
   * Callback for stdout messages.
   */
  onStdout?: (message: OutputMessage) => any | Promise<any>
  /**
   * Callback for stderr messages.
   */
  onStderr?: (message: OutputMessage) => any | Promise<any>
  /**
   * Callback for execution errors (e.g., runtime exceptions).
   */
  onError?: (error: ExecutionError) => any | Promise<any>
}
