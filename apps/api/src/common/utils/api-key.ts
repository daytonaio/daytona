/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import * as crypto from 'crypto'

export function generateRandomString(size: number): string {
  return crypto.randomBytes(size).toString('hex')
}

export function generateApiKeyValue(): string {
  return `dtn_${generateRandomString(32)}`
}

export function generateApiKeyHash(value: string): string {
  return crypto.createHash('sha256').update(value).digest('hex')
}
