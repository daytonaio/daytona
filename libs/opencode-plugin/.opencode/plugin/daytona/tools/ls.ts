/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { ToolContext, PluginInput } from '@opencode-ai/plugin'
import type { DaytonaSessionManager } from '../core/session-manager'
import type { FileInfo } from '@daytonaio/sdk'

export const lsTool = (sessionManager: DaytonaSessionManager, projectId: string, worktree: string, pluginCtx: PluginInput) => ({
  description: 'Lists files in a directory in Daytona sandbox',
  args: {
    dirPath: z.string().optional(),
  },
  async execute(args: { dirPath?: string }, ctx: ToolContext) {
    const sandbox = await sessionManager.getSandbox(ctx.sessionID, projectId, worktree, pluginCtx)
    const workDir = await sandbox.getWorkDir()
    const path = args.dirPath || workDir
    if (!path) {
      throw new Error('Work directory not available')
    }
    const files = (await sandbox.fs.listFiles(path)) as FileInfo[]
    return files.map((f) => f.name).join('\n')
  },
})
