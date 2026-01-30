/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { PluginInput } from '@opencode-ai/plugin'
import type { ToolContext } from '@opencode-ai/plugin/tool'
import type { DaytonaSessionManager } from '../core/session-manager'

export const editTool = (
  sessionManager: DaytonaSessionManager,
  projectId: string,
  worktree: string,
  pluginCtx: PluginInput,
) => ({
  description: 'Replaces text in a file in Daytona sandbox',
  args: {
    filePath: z.string(),
    oldString: z.string(),
    newString: z.string(),
  },
  async execute(args: { filePath: string; oldString: string; newString: string }, ctx: ToolContext) {
    const sandbox = await sessionManager.getSandbox(ctx.sessionID, projectId, worktree, pluginCtx)
    const buffer = await sandbox.fs.downloadFile(args.filePath)
    const decoder = new TextDecoder()
    const content = decoder.decode(buffer)
    const newContent = content.replace(args.oldString, args.newString)
    await sandbox.fs.uploadFile(Buffer.from(newContent), args.filePath)
    return `Edited ${args.filePath}`
  },
})
