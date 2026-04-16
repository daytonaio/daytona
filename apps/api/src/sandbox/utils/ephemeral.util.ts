/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export function isEphemeral(sandbox: { autoDeleteInterval?: number }): boolean {
  return sandbox.autoDeleteInterval === 0
}
