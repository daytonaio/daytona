/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Sandbox-side git push.
 *
 * The agent commits its own work inside the sandbox (it's prompted to commit but
 * not push). The extension only pushes those commits to the session's GitHub
 * branch via the Daytona git API, using a transient token as the credential.
 * Pushes are serialized so overlapping triggers can't race, and a branch with
 * nothing ahead of its remote is skipped.
 */

import type { Sandbox } from '@daytona/sdk'
import { withRecovery } from './sandbox.ts'

const PUSH_USERNAME = 'x-access-token'

export interface PushTarget {
  sandbox: Sandbox
  cwd: string
  /** True when a GitHub origin and token are available (otherwise a no-op). */
  pushEnabled: boolean
}

export interface PushResult {
  pushed: boolean
}

// Serialize pushes across the whole extension so concurrent triggers don't race.
let queue: Promise<unknown> = Promise.resolve()

export function pushChanges(target: PushTarget, token: string | undefined): Promise<PushResult> {
  const next = queue.then(() => doPush(target, token))
  queue = next.catch(() => {
    /* keep the chain alive even if one push throws */
  })
  return next
}

async function doPush(target: PushTarget, token: string | undefined): Promise<PushResult> {
  if (!target.pushEnabled || !token) return { pushed: false }

  const { sandbox, cwd } = target
  const status = await withRecovery(sandbox, () => sandbox.git.status(cwd))
  // Nothing to push: no local commits ahead of the remote branch.
  if ((status.ahead ?? 0) <= 0) return { pushed: false }

  await withRecovery(sandbox, () => sandbox.git.push(cwd, PUSH_USERNAME, token))
  return { pushed: true }
}
