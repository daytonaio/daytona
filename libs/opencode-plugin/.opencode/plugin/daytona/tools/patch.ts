/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { ToolContext, PluginInput } from '@opencode-ai/plugin'
import type { DaytonaSessionManager } from '../core/session-manager'

export const patchTool = (
  sessionManager: DaytonaSessionManager,
  projectId: string,
  worktree: string,
  pluginCtx: PluginInput,
) => ({
  description: 'Patches a file with a code snippet in Daytona sandbox',
  args: {
    filePath: z.string(),
    oldSnippet: z.string(),
    newSnippet: z.string(),
  },
  async execute(args: { filePath: string; oldSnippet: string; newSnippet: string }, ctx: ToolContext) {
    const sandbox = await sessionManager.getSandbox(ctx.sessionID, projectId, worktree, pluginCtx)
    const buffer = await sandbox.fs.downloadFile(args.filePath)
    const decoder = new TextDecoder()
    const content = decoder.decode(buffer)
    const newContent = content.replace(args.oldSnippet, args.newSnippet)
    await sandbox.fs.uploadFile(Buffer.from(newContent), args.filePath)
    return `Patched ${args.filePath}`
  },
})
