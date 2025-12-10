/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox } from '../entities/sandbox.entity'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { RunnerService } from '../services/runner.service'

/**
 * Checks if a sandbox is recoverable based on the error reason and updates the sandbox.recoverable property
 * @param sandbox - The sandbox entity to check and update
 * @param runnerService - The runner service to fetch the runner
 * @param runnerAdapterFactory - The runner adapter factory to create the adapter
 */
export async function checkRecoverable(
  sandbox: Sandbox,
  runnerService: RunnerService,
  runnerAdapterFactory: RunnerAdapterFactory,
): Promise<void> {
  sandbox.recoverable = false

  if (!sandbox.errorReason || !sandbox.runnerId) {
    return
  }

  try {
    const runner = await runnerService.findOne(sandbox.runnerId)
    const runnerAdapter = await runnerAdapterFactory.create(runner)
    sandbox.recoverable = await runnerAdapter.isRecoverable(sandbox.id, sandbox.errorReason)
  } catch {
    // Keep sandbox.recoverable = false (already set)
  }
}
