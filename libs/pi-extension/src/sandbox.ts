/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Resilience layer for sandbox operations.
 *
 * A sandbox can become unavailable mid-session: Daytona auto-stops it after an
 * idle period, or it gets reaped/removed. A stopped sandbox is recoverable
 * (its filesystem is preserved) — we just need to start it again. A deleted one
 * is gone for good.
 *
 * `withRecovery` runs an operation and, on failure, inspects the real sandbox
 * state (rather than matching error strings): if it's merely stopped, it starts
 * it and retries once; if it's gone, it throws a clear, actionable error
 * instead of a cryptic Docker "no such container" message. Tool execution then
 * fails loudly — it never silently falls back to the host.
 */

import type { Sandbox } from '@daytona/sdk'

export class SandboxUnavailableError extends Error {
  constructor(public readonly sandboxId: string) {
    super(
      `Daytona sandbox ${sandboxId.slice(0, 8)} is no longer available — it was likely ` +
        `reaped after inactivity or removed. Tool execution is paused (it was NOT run ` +
        `locally). Restart Pi with --daytona to get a fresh sandbox.`,
    )
    this.name = 'SandboxUnavailableError'
  }
}

/**
 * Run a sandbox operation, transparently restarting a stopped sandbox and
 * retrying once. Throws SandboxUnavailableError if the sandbox is gone.
 */
export async function withRecovery<T>(sandbox: Sandbox, fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    // Figure out whether this is an availability problem or a real op error.
    let state: string | undefined
    try {
      await sandbox.refreshData()
      state = sandbox.state
    } catch {
      // refreshData failing means the sandbox no longer exists.
      throw new SandboxUnavailableError(sandbox.id)
    }
    if (state !== undefined && state !== 'started') {
      // Stopped/archived but still present: bring it back and retry once.
      try {
        await sandbox.start()
      } catch {
        throw new SandboxUnavailableError(sandbox.id)
      }
      return await fn()
    }
    // Sandbox is running fine — this was a genuine operation error.
    throw err
  }
}

/** executeCommand with auto-recovery. */
export async function execCommand(sandbox: Sandbox, command: string, cwd?: string, timeout?: number) {
  return withRecovery(sandbox, () => sandbox.process.executeCommand(command, cwd, undefined, timeout))
}
