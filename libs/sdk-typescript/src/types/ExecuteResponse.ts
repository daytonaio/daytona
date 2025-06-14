/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ExecuteResponse as ClientExecuteResponse } from '@daytonaio/api-client'
import { Chart } from './Charts'

/**
 * Artifacts from the command execution.
 *
 * @interface
 * @property stdout - Standard output from the command, same as `result` in `ExecuteResponse`
 * @property charts - List of chart metadata from matplotlib
 */
export interface ExecutionArtifacts {
  stdout: string
  charts?: Chart[]
}

/**
 * Response from the command execution.
 *
 * @interface
 * @property exitCode - The exit code from the command execution
 * @property result - The output from the command execution
 * @property artifacts - Artifacts from the command execution
 */
export interface ExecuteResponse extends ClientExecuteResponse {
  exitCode: number
  result: string
  artifacts?: ExecutionArtifacts
}
