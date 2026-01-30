/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { PluginInput } from '@opencode-ai/plugin'
import type { ToolContext } from '@opencode-ai/plugin/tool'
import type { DaytonaSessionManager } from '../core/session-manager'

export const multieditTool = (
  sessionManager: DaytonaSessionManager,
  projectId: string,
  worktree: string,
  pluginCtx: PluginInput,
) => ({
  description: 'Applies multiple edits to a file in Daytona sandbox atomically',
  args: {
    filePath: z.string(),
    edits: z.array(
      z.object({
        oldString: z.string(),
        newString: z.string(),
      }),
    ),
  },
  async execute(args: { filePath: string; edits: Array<{ oldString: string; newString: string }> }, ctx: ToolContext) {
    const sandbox = await sessionManager.getSandbox(ctx.sessionID, projectId, worktree, pluginCtx)
    const buffer = await sandbox.fs.downloadFile(args.filePath)
    const decoder = new TextDecoder()
    let content = decoder.decode(buffer)

    for (const edit of args.edits) {
      content = content.replace(edit.oldString, edit.newString)
    }

    await sandbox.fs.uploadFile(Buffer.from(content), args.filePath)
    return `Applied ${args.edits.length} edits to ${args.filePath}`
  },
})
