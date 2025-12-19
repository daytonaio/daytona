/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Response } from 'express'

/**
 * Utility functions for setting rate limit headers consistently across middleware and services
 */

export interface RateLimitHeadersOptions {
  throttlerName: string
  limit: number
  remaining: number
  resetSeconds: number
  retryAfterSeconds?: number
}

/**
 * Sets standard rate limit headers on a response
 * Follows the pattern: X-RateLimit-{Limit|Remaining|Reset}-{throttlerName}
 */
export function setRateLimitHeaders(response: Response, options: RateLimitHeadersOptions): void {
  const { throttlerName, limit, remaining, resetSeconds, retryAfterSeconds } = options

  response.setHeader(`X-RateLimit-Limit-${throttlerName}`, limit.toString())
  response.setHeader(`X-RateLimit-Remaining-${throttlerName}`, remaining.toString())
  response.setHeader(`X-RateLimit-Reset-${throttlerName}`, resetSeconds.toString())

  if (retryAfterSeconds !== undefined) {
    response.setHeader('Retry-After', retryAfterSeconds.toString())
  }
}
