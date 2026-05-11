/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Direct-to-runner access bundle for the in-sandbox session-daemon. The URL already encodes
 * `${runner.apiUrl}/sandboxes/<id>/toolbox/proxy/<daemonPort>` so callers just append the
 * daemon's path (e.g. `/sessions`, `/load`). This bypasses the public proxy on port 4000 for
 * API-internal calls — one fewer TCP hop and auth round-trip per daemon call.
 */
export interface DaemonAccess {
  url: string
  runnerApiKey: string
}

/**
 * Build the API-internal daemon access bundle from a resolved runner. Shared by SessionService
 * (exec/create) and SessionLoadService (load polling) so the proxy URL shape lives in one place.
 */
export function buildDaemonAccess(
  runner: { apiUrl?: string; apiKey?: string } | null | undefined,
  sandboxId: string,
  port: number,
): DaemonAccess {
  if (!runner || !runner.apiUrl || !runner.apiKey) {
    throw new Error(`runner for sandbox ${sandboxId} is missing apiUrl/apiKey`)
  }
  return {
    url: `${runner.apiUrl.replace(/\/$/, '')}/sandboxes/${sandboxId}/toolbox/proxy/${port}`,
    runnerApiKey: runner.apiKey,
  }
}
