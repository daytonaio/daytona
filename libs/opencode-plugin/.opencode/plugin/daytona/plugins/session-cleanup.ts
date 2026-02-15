/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Plugin, PluginInput } from '@opencode-ai/plugin'
import type { DaytonaSessionManager } from '../core/session-manager'
import { EVENT_TYPE_SESSION_DELETED } from '../core/types'
import type { EventSessionDeleted } from '../core/types'
import { toast } from '../core/toast'

/**
 * Creates the session cleanup plugin for Daytona
 * Automatically cleans up sandbox resources when sessions end
 */
export function createSessionCleanupPlugin(sessionManager: DaytonaSessionManager): Plugin {
  return async (pluginCtx: PluginInput) => {
    toast.initialize(pluginCtx.client?.tui)
    const projectId = pluginCtx.project.id
    return {
      event: async ({ event }) => {
        if (event.type === EVENT_TYPE_SESSION_DELETED) {
          const sessionId = (event as EventSessionDeleted).properties.info.id
          try {
            await sessionManager.deleteSandbox(sessionId, projectId)
            toast.show({
              title: 'Session deleted',
              message: 'Sandbox deleted successfully.',
              variant: 'success',
            })
          } catch (err: any) {
            toast.show({
              title: 'Delete failed',
              message: err?.message || 'Failed to delete sandbox.',
              variant: 'error',
            })
            throw err
          }
        }
      },
    }
  }
}
