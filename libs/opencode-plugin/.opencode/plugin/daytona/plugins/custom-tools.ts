/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { PluginInput } from '@opencode-ai/plugin'
import { createDaytonaTools } from '../tools'
import { logger } from '../core/logger'
import type { DaytonaSessionManager } from '../core/session-manager'

/**
 * Custom tools for Daytona sandbox: file ops, command execution, search.
 */
export async function customTools(ctx: PluginInput, sessionManager: DaytonaSessionManager) {
  logger.info('OpenCode started with Daytona plugin')
  const projectId = ctx.project.id
  const worktree = ctx.project.worktree
  return createDaytonaTools(sessionManager, projectId, worktree, ctx)
}
