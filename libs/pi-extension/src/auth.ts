/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/** API key resolution: env var first, one-time UI prompt as fallback. */

import type { ExtensionContext } from '@earendil-works/pi-coding-agent'

/**
 * Resolve the Daytona API key.
 *
 * Order:
 *   1. `DAYTONA_API_KEY` environment variable (the documented SDK convention).
 *   2. A one-time interactive prompt for this session (when a UI is available).
 *
 * There is no extension-writable secrets vault, so a prompted key is held only
 * in memory for the session and never persisted.
 */
export async function resolveApiKey(ctx: ExtensionContext): Promise<string | undefined> {
  const fromEnv = process.env.DAYTONA_API_KEY?.trim()
  if (fromEnv) return fromEnv

  if (!ctx.hasUI) return undefined

  const entered = await ctx.ui.input('Daytona API key', 'DAYTONA_API_KEY — create one at https://app.daytona.io')
  return entered?.trim() || undefined
}
