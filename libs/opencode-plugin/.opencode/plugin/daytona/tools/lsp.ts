/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { ToolContext, PluginInput } from '@opencode-ai/plugin'
import type { DaytonaSessionManager } from '../core/session-manager'

export const lspTool = (
  sessionManager: DaytonaSessionManager,
  projectId: string,
  worktree: string,
  pluginCtx: PluginInput,
) => ({
  description: 'LSP operation in Daytona sandbox (code intelligence)',
  args: {
    op: z.string(),
    filePath: z.string(),
    line: z.number(),
  },
  async execute(args: { op: string; filePath: string; line: number }, ctx: ToolContext) {
    return `LSP operations are not yet implemented in the Daytona plugin.`
  },
})
