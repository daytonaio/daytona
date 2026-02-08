/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { readFileSync, existsSync } from 'fs'
import dotenv from 'dotenv'

// Merge two objects recursively
export function deepMerge<T>(target: T, source: Record<string, unknown>): T {
  const out = { ...target } as Record<string, unknown>
  for (const key of Object.keys(source)) {
    const a = (out as Record<string, unknown>)[key]
    const b = source[key]
    if (
      a != null &&
      b != null &&
      typeof a === 'object' &&
      typeof b === 'object' &&
      !Array.isArray(a) &&
      !Array.isArray(b)
    ) {
      ;(out as Record<string, unknown>)[key] = deepMerge(a as Record<string, unknown>, b as Record<string, unknown>)
    } else {
      ;(out as Record<string, unknown>)[key] = b
    }
  }
  return out as T
}

// Read env file and return a record of key-value pairs
export function readEnvFile(path: string): Record<string, string> {
  if (!existsSync(path)) return {}
  const parsed = dotenv.parse(readFileSync(path))
  return Object.fromEntries(
    Object.entries(parsed).filter(([, v]) => v != null && v !== '') as [string, string][],
  ) as Record<string, string>
}
