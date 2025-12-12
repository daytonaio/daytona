/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox } from '../entities/sandbox.entity'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { RunnerService } from '../services/runner.service'

/**
 * Checks if a sandbox is recoverable based on the error reason.
 * @param sandbox - The sandbox entity to check
 * @param runnerService - The runner service to fetch the runner
 * @param runnerAdapterFactory - The runner adapter factory to create the adapter
 * @returns Promise<boolean> - true if recoverable, false otherwise
 */
export async function checkRecoverable(
  sandbox: Sandbox,
  runnerService: RunnerService,
  runnerAdapterFactory: RunnerAdapterFactory,
): Promise<boolean> {
  if (!sandbox.errorReason || !sandbox.runnerId) {
    return false
  }

  try {
    const runner = await runnerService.findOne(sandbox.runnerId)
    const runnerAdapter = await runnerAdapterFactory.create(runner)
    return await runnerAdapter.isRecoverable(sandbox.id, sandbox.errorReason)
  } catch {
    // If an error occurs, consider not recoverable
    return false
  }
}
