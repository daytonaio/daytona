/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { ToolContext, PluginInput } from '@opencode-ai/plugin'
import type { DaytonaSessionManager } from '../core/session-manager'

export const globTool = (sessionManager: DaytonaSessionManager, projectId: string, worktree: string, pluginCtx: PluginInput) => ({
  description: 'Searches for files matching a pattern in Daytona sandbox',
  args: {
    pattern: z.string(),
  },
  async execute(args: { pattern: string }, ctx: ToolContext) {
    const sandbox = await sessionManager.getSandbox(ctx.sessionID, projectId, worktree, pluginCtx)
    const workDir = await sandbox.getWorkDir()
    if (!workDir) {
      throw new Error('Work directory not available')
    }
    const result = await sandbox.fs.searchFiles(workDir, args.pattern)
    return result.files.join('\n')
  },
})
