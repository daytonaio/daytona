/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { z } from 'zod'
import type { PluginInput } from '@opencode-ai/plugin'
import type { ToolContext } from '@opencode-ai/plugin/tool'
import type { DaytonaSessionManager } from '../core/session-manager'

export const patchTool = (
  sessionManager: DaytonaSessionManager,
  projectId: string,
  worktree: string,
  pluginCtx: PluginInput,
) => ({
  description: 'Applies a patch to the project in Daytona sandbox',
  args: {
    patchText: z.string().describe('The full patch text that describes all changes to be made'),
  },
  async execute(args: { filePath: string; oldSnippet: string; newSnippet: string }, ctx: ToolContext) {
    return `Patch operations are not yet implemented in the Daytona plugin.`
  },
})
