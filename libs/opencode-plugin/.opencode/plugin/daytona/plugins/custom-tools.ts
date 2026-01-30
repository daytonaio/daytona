/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Plugin, PluginInput } from '@opencode-ai/plugin'
import { createDaytonaTools } from '../tools'
import { logger } from '../core/logger'
import type { DaytonaSessionManager } from '../core/session-manager'
import { toast } from '../core/toast'

/**
 * Creates the custom tools plugin for Daytona sandbox integration
 * Provides tools for file operations, command execution, and search within sandboxes
 */
export function createCustomToolsPlugin(sessionManager: DaytonaSessionManager): Plugin {
  return async (pluginCtx: PluginInput) => {
    logger.info('OpenCode started with Daytona plugin')
    toast.initialize(pluginCtx.client?.tui)

    const projectId = pluginCtx.project.id
    const worktree = pluginCtx.project.worktree

    return {
      tool: createDaytonaTools(sessionManager, projectId, worktree, pluginCtx),
    }
  }
}
