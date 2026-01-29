/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Type definitions and constants for the Daytona OpenCode plugin
 */

import type { Sandbox } from '@daytonaio/sdk'

// OpenCode Types

export type EventSessionDeleted = {
  type: 'session.deleted'
  properties: {
    info: { id: string }
  }
}

export type EventSessionIdle = {
  type: 'session.idle'
  properties: {
    sessionID: string
  }
}

export type ExperimentalChatSystemTransformInput = {
  sessionID: string
}

export type ExperimentalChatSystemTransformOutput = {
  system: string[]
}

// OpenCode constants

export const EVENT_TYPE_SESSION_DELETED = 'session.deleted'
export const EVENT_TYPE_SESSION_IDLE = 'session.idle'

// Daytona plugin types

export type LogLevel = 'INFO' | 'ERROR' | 'WARN'

export type SandboxInfo = {
  id: string
}

export type SessionInfo = {
  sandboxId: string
  /**
   * Only set when the local worktree is a git repo (used to create opencode/N branches/remotes).
   */
  branchNumber?: number
  created: number
  lastAccessed: number
}

export type ProjectSessionData = {
  projectId: string
  worktree: string
  sessions: Record<string, SessionInfo>
}

export type SessionSandboxMap = Map<string, Sandbox | SandboxInfo>

// Daytona plugin constants

export const LOG_LEVEL_INFO: LogLevel = 'INFO'
export const LOG_LEVEL_ERROR: LogLevel = 'ERROR'
export const LOG_LEVEL_WARN: LogLevel = 'WARN'
