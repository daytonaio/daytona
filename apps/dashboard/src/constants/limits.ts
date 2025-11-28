/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const TIER_REQUIREMENTS: Record<number, string[]> = {
  1: ['Email verification'],
  2: ['Credit card linked', 'GitHub connected', 'Top up $25 (one time)'],
  3: ['Business email verified', 'Top up $500 (one time)'],
  4: ['Top up $2,000 (every 30 days)'],
}

export const TIER_RATE_LIMITS: Record<
  number,
  { authenticatedRateLimit: number; sandboxCreateRateLimit: number; sandboxLifecycleRateLimit: number }
> = {
  1: {
    authenticatedRateLimit: 40_000,
    sandboxCreateRateLimit: 500,
    sandboxLifecycleRateLimit: 40_000,
  },
  2: {
    authenticatedRateLimit: 100_000,
    sandboxCreateRateLimit: 1000,
    sandboxLifecycleRateLimit: 100_000,
  },
  3: {
    authenticatedRateLimit: 200_000,
    sandboxCreateRateLimit: 2000,
    sandboxLifecycleRateLimit: 200_000,
  },
  4: {
    authenticatedRateLimit: 400_000,
    sandboxCreateRateLimit: 4000,
    sandboxLifecycleRateLimit: 400_000,
  },
}
