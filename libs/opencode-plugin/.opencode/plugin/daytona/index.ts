/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * OpenCode Plugin: Daytona Sandbox Integration
 *
 * OpenCode plugins extend the AI coding assistant by adding custom tools, handling events,
 * and modifying behavior. Plugins are TypeScript/JavaScript modules that export functions
 * which return hooks for various lifecycle events.
 *
 * This plugin integrates Daytona sandboxes with OpenCode, providing isolated development
 * environments for each session. It adds custom tools for file operations, command execution,
 * and search within sandboxes, and automatically cleans up resources when sessions end.
 *
 * Learn more: https://opencode.ai/docs/plugins/
 *
 * Daytona Sandbox Integration Tools
 *
 * Requires:
 * - npm install @daytonaio/sdk
 * - Environment: DAYTONA_API_KEY
 */

import { join } from 'path'
import { xdgData } from 'xdg-basedir'
import type { Plugin } from '@opencode-ai/plugin'

// Import modules
import { setLogFilePath } from './core/logger'
import { DaytonaSessionManager } from './core/session-manager'
import {
  createCustomToolsPlugin,
  createSessionCleanupPlugin,
  createSystemTransformPlugin,
  createSessionIdleAutoCommitPlugin,
} from './plugins'

// Export types for consumers
export type { EventSessionDeleted, LogLevel, SandboxInfo, SessionInfo, ProjectSessionData } from './core/types'

// Initialize logger and session manager using xdg-basedir (same as OpenCode)
const LOG_FILE = join(xdgData!, 'opencode', 'log', 'daytona.log')
const STORAGE_DIR = join(xdgData!, 'opencode', 'storage', 'daytona')
const REPO_PATH = '/home/daytona/project'

setLogFilePath(LOG_FILE)
const sessionManager = new DaytonaSessionManager(process.env.DAYTONA_API_KEY || '', STORAGE_DIR, REPO_PATH)

// Export plugin instances
export const CustomToolsPlugin: Plugin = createCustomToolsPlugin(sessionManager)
export const DaytonaSessionCleanupPlugin: Plugin = createSessionCleanupPlugin(sessionManager)
export const SystemTransformPlugin: Plugin = createSystemTransformPlugin(REPO_PATH)
export const DaytonaSessionIdleAutoCommitPlugin: Plugin = createSessionIdleAutoCommitPlugin(sessionManager, REPO_PATH)
