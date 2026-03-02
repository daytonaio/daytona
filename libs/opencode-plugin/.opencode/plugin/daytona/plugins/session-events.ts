/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { PluginInput } from '@opencode-ai/plugin'
import { SessionGitManager } from '../git/session-git-manager'
import { EVENT_TYPE_SESSION_DELETED, EVENT_TYPE_SESSION_IDLE, type EventSessionDeleted } from '../core/types'
import { toast } from '../core/toast'
import { logger } from '../core/logger'
import type { DaytonaSessionManager } from '../core/session-manager'

/**
 * Handles OpenCode session events.
 */
export async function eventHandlers(ctx: PluginInput, sessionManager: DaytonaSessionManager, repoPath: string) {
  const projectId = ctx.project.id
  const worktree = ctx.project.worktree
  return async (args: any) => {
    const event = args.event
    if (event.type === EVENT_TYPE_SESSION_DELETED) {
      const sessionId = (event as EventSessionDeleted).properties.info.id
      try {
        await sessionManager.deleteSandbox(sessionId, projectId)
        toast.show({ title: 'Session deleted', message: 'Sandbox deleted successfully.', variant: 'success' })
      } catch (err: any) {
        toast.show({ title: 'Delete failed', message: err?.message || 'Failed to delete sandbox.', variant: 'error' })
        throw err
      }
    } else if (event.type === EVENT_TYPE_SESSION_IDLE) {
      const sessionId = event.properties.sessionID
      const start = Date.now()
      try {
        const sandbox = await sessionManager.getSandbox(sessionId, projectId, worktree, ctx)
        const branchNumber = sessionManager.getBranchNumberForSandbox(projectId, sandbox.id)
        if (!branchNumber) return
        const sessionGit = new SessionGitManager(sandbox, repoPath, worktree, branchNumber)
        const didSync = await sessionGit.autoCommitAndPull(ctx)
        logger.info(
          `[idle] done sessionId=${sessionId} sandboxId=${sandbox.id} synced=${didSync} in ${Date.now() - start}ms`,
        )
      } catch (err: any) {
        logger.error(`[idle] error sessionId=${sessionId} in ${Date.now() - start}ms: ${err}`)
        toast.show({
          title: 'Auto-commit error',
          message: err?.message || 'Failed to auto-commit and pull.',
          variant: 'error',
        })
        throw err
      }
    }
  }
}
