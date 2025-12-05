/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import * as crypto from 'crypto'

export function generateApiKeyValue(): string {
  return `dtn_${crypto.randomBytes(32).toString('hex')}`
}

export function generateApiKeyHash(value: string): string {
  return crypto.createHash('sha256').update(value).digest('hex')
}
