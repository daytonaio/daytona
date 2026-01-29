/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Tool implementations for Daytona sandbox integration
 */

import { bashTool } from './tools/bash'
import { readTool } from './tools/read'
import { writeTool } from './tools/write'
import { editTool } from './tools/edit'
import { multieditTool } from './tools/multiedit'
import { patchTool } from './tools/patch'
import { lsTool } from './tools/ls'
import { globTool } from './tools/glob'
import { grepTool } from './tools/grep'
import { lspTool } from './tools/lsp'
import { getPreviewURLTool } from './tools/get-preview-url'

import type { DaytonaSessionManager } from './core/session-manager'
import type { PluginInput } from '@opencode-ai/plugin'

export function createDaytonaTools(
  sessionManager: DaytonaSessionManager,
  projectId: string,
  worktree: string,
  pluginCtx: PluginInput,
) {
  const repoPath = sessionManager.repoPath
  return {
    bash: bashTool(sessionManager, projectId, worktree, pluginCtx, repoPath),
    read: readTool(sessionManager, projectId, worktree, pluginCtx),
    write: writeTool(sessionManager, projectId, worktree, pluginCtx),
    edit: editTool(sessionManager, projectId, worktree, pluginCtx),
    multiedit: multieditTool(sessionManager, projectId, worktree, pluginCtx),
    patch: patchTool(sessionManager, projectId, worktree, pluginCtx),
    ls: lsTool(sessionManager, projectId, worktree, pluginCtx),
    glob: globTool(sessionManager, projectId, worktree, pluginCtx),
    grep: grepTool(sessionManager, projectId, worktree, pluginCtx),
    lsp: lspTool(sessionManager, projectId, worktree, pluginCtx),
    getPreviewURL: getPreviewURLTool(sessionManager, projectId, worktree, pluginCtx),
  }
}
