/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

// List of substrings that indicate a recoverable error
export const RECOVERABLE_ERRORS = [
  'no space left on device',
  'storage limit',
  // add more as needed
]

export function isRecoverableError(reason?: string): boolean {
  if (!reason) return false
  const msg = reason.toLowerCase()
  return RECOVERABLE_ERRORS.some((substr) => msg.includes(substr))
}
