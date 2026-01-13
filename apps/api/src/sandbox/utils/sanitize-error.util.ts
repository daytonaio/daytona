/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export function sanitizeSandboxError(error: any): { recoverable: boolean; errorReason: string } {
  if (typeof error === 'string') {
    try {
      const errObj = JSON.parse(error) as { recoverable: boolean; errorReason: string }
      return { recoverable: errObj.recoverable, errorReason: errObj.errorReason }
    } catch {
      return { recoverable: false, errorReason: error }
    }
  } else if (typeof error === 'object' && error !== null && 'recoverable' in error && 'errorReason' in error) {
    return { recoverable: error.recoverable, errorReason: error.errorReason }
  } else if (typeof error === 'object' && error.message) {
    return sanitizeSandboxError(error.message)
  }

  return { recoverable: false, errorReason: String(error) }
}
