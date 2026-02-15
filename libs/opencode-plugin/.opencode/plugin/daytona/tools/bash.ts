/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { PluginInput } from '@opencode-ai/plugin'
import type { ToolContext } from '@opencode-ai/plugin/tool'
import type { DaytonaSessionManager } from '../core/session-manager'

export const bashTool = (
  sessionManager: DaytonaSessionManager,
  projectId: string,
  worktree: string,
  pluginCtx: PluginInput,
  repoPath: string,
) => ({
  description: 'Executes shell commands in a Daytona sandbox',
  args: {
    command: z.string(),
    background: z.boolean().optional(),
  },
  async execute(args: { command: string; background?: boolean }, ctx: ToolContext) {
    const sessionId = ctx.sessionID
    const sandbox = await sessionManager.getSandbox(sessionId, projectId, worktree, pluginCtx)

    if (args.background) {
      const execSessionId = `exec-session-${sessionId}`
      try {
        await sandbox.process.getSession(execSessionId)
      } catch {
        await sandbox.process.createSession(execSessionId)
      }
      await sandbox.process.executeSessionCommand(execSessionId, {
        command: `cd ${repoPath}`,
      })
      const result = await sandbox.process.executeSessionCommand(execSessionId, {
        command: args.command,
        runAsync: true,
      })
      return `Command started in background (cmdId: ${result.cmdId})`
    } else {
      const result = await sandbox.process.executeCommand(args.command, repoPath)
      return `Exit code: ${result.exitCode}\n${result.result}`
    }
  },
})
