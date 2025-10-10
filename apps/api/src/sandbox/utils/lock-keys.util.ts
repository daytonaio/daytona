/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Centralized lock key definitions for sandbox operations.
 * This ensures consistent locking across all sandbox state changes.
 */

/**
 * Generates a lock key for sandbox state-changing operations.
 * This lock ensures mutual exclusion across all operations that modify
 * sandbox state (start, stop, destroy, archive, auto-stop, auto-delete, etc.)
 *
 * @param sandboxId - The ID of the sandbox
 * @returns Lock key in format: sandbox:{sandboxId}:state-change
 */
export function getSandboxStateChangeLockKey(sandboxId: string): string {
  return `sandbox:${sandboxId}:state-change`
}
