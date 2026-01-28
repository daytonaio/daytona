import type { Plugin, PluginInput } from '@opencode-ai/plugin'
import type { DaytonaSessionManager } from '../core/session-manager'
import { SessionGitManager } from '../git/session-git-manager'
import { EVENT_TYPE_SESSION_IDLE } from '../core/types'
import { toast } from '../core/toast'

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
          try {
            const sandbox = await sessionManager.getSandbox(sessionId, projectId, worktree, pluginCtx)
            const branchNumber = sessionManager.getBranchNumberForSandbox(projectId, sandbox.id)
            if (!branchNumber) {
              toast.show({
                title: 'Auto-commit failed',
                message: `No branch number found for sandbox ${sandbox.id}`,
                variant: 'error',
              })
              throw new Error(`No branch number found for sandbox ${sandbox.id}`)
            }
            const sessionGit = new SessionGitManager(sandbox, repoPath, branchNumber)
            await sessionGit.autoCommitAndPull(pluginCtx)
          } catch (err: any) {
            toast.show({
              title: 'Auto-commit error',
              message: err?.message || 'Failed to auto-commit and pull.',
              variant: 'error',
            })
            throw err
          }
        }
      },
    }
  }
}
