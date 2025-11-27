/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Recovery types for sandbox errors
 */
export enum RecoveryType {
  STORAGE_EXPANSION = 'storage-expansion',
}

/**
 * Patterns that indicate recoverable errors
 */
const RECOVERABLE_ERROR_PATTERNS = {
  [RecoveryType.STORAGE_EXPANSION]: [/no space left on device/i, /storage limit/i, /disk quota exceeded/i],
} as const

/**
 * Determines if an error reason indicates a recoverable error
 * @param errorReason - The sandbox error reason string
 * @returns The recovery type if recoverable, null otherwise
 */
export function detectRecoveryType(errorReason: string | null | undefined): RecoveryType | null {
  if (!errorReason) return null

  for (const [recoveryType, patterns] of Object.entries(RECOVERABLE_ERROR_PATTERNS)) {
    if (patterns.some((pattern) => pattern.test(errorReason))) {
      return recoveryType as RecoveryType
    }
  }

  return null
}

/**
 * Checks if an error reason is recoverable (any type)
 * @param errorReason - The sandbox error reason string
 * @returns true if recoverable, false otherwise
 */
export function isRecoverableError(errorReason: string | null | undefined): boolean {
  return detectRecoveryType(errorReason) !== null
}
