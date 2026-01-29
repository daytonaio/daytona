/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { ToolContext, PluginInput } from '@opencode-ai/plugin'
import type { DaytonaSessionManager } from '../core/session-manager'

export const getPreviewURLTool = (sessionManager: DaytonaSessionManager, projectId: string, worktree: string, pluginCtx: PluginInput) => ({
  description: 'Gets a preview URL for the Daytona sandbox',
  args: {
    port: z.number(),
  },
  async execute(args: { port: number }, ctx: ToolContext) {
    const sandbox = await sessionManager.getSandbox(ctx.sessionID, projectId, worktree, pluginCtx)
    const previewLink = await sandbox.getPreviewLink(args.port)
    return `Sandbox Preview URL: ${previewLink.url}`
  },
})
