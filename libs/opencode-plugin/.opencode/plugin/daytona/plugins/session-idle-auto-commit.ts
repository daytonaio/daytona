/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Plugin, PluginInput } from '@opencode-ai/plugin'
import type { DaytonaSessionManager } from '../core/session-manager'
import { SessionGitManager } from '../git/session-git-manager'
import { EVENT_TYPE_SESSION_IDLE } from '../core/types'
import { toast } from '../core/toast'
import { logger } from '../core/logger'

/**
 * Creates a plugin to auto-commit in the sandbox on session idle
 */
export function createSessionIdleAutoCommitPlugin(sessionManager: DaytonaSessionManager, repoPath: string): Plugin {
  return async (pluginCtx: PluginInput) => {
    toast.initialize(pluginCtx.client?.tui)
    const projectId = pluginCtx.project.id
    const worktree = pluginCtx.project.worktree

    return {
      event: async (args: any) => {
        const event = args.event
        if (event.type === EVENT_TYPE_SESSION_IDLE) {
          const sessionId = event.properties.sessionID
          const start = Date.now()
          try {
            const sandbox = await sessionManager.getSandbox(sessionId, projectId, worktree, pluginCtx)
            const branchNumber = sessionManager.getBranchNumberForSandbox(projectId, sandbox.id)
            if (!branchNumber) {
              // No local git repo => no branch reservation => nothing to sync.
              return
            }
            const sessionGit = new SessionGitManager(sandbox, repoPath, worktree, branchNumber)
            const didSync = await sessionGit.autoCommitAndPull(pluginCtx)
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
          } finally {
            // Intentionally no-op; keep logs minimal.
          }
        }
      },
    }
  }
}
