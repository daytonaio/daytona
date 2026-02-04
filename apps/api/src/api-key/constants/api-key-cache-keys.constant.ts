/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export function getApiKeyValidationCacheKey(keyHash: string): string {
  return `api-key:validation:${keyHash}`
}

export function getApiKeyUserCacheKey(userId: string): string {
  return `api-key:user:${userId}`
}
