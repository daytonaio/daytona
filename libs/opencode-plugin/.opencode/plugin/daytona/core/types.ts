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
  branchNumber: number
  created: number
  lastAccessed: number
}

export type ProjectSessionData = {
  projectId: string
  worktree: string
  /**
   * Monotonically increasing pointer for branch numbering.
   * We persist this so we don't reuse branch numbers after sessions are deleted.
   */
  lastBranchNumber?: number
  sessions: Record<string, SessionInfo>
}

export type SessionSandboxMap = Map<string, Sandbox | SandboxInfo>

// Daytona plugin constants

export const LOG_LEVEL_INFO: LogLevel = 'INFO'
export const LOG_LEVEL_ERROR: LogLevel = 'ERROR'
export const LOG_LEVEL_WARN: LogLevel = 'WARN'
