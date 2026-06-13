/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Type definitions and constants for the Daytona OpenCode plugin
 */

import type { Sandbox } from '@daytona/sdk'

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
  sessionID?: string
  model: any
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

export type SandboxCreationParams =
  | {
      image: string
    }
  | {
      snapshot: string
    }

export type DaytonaPluginOptions = Record<string, unknown> & {
  image?: unknown
  snapshot?: unknown
}

function normalizeOptionalString(value: unknown): string | undefined {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed ? trimmed : undefined
}

export function resolveSandboxCreationParams(options?: DaytonaPluginOptions): SandboxCreationParams | undefined {
  const pluginOptions = (options ?? {}) as DaytonaPluginOptions
  const configuredImage = normalizeOptionalString(pluginOptions.image)
  const configuredSnapshot = normalizeOptionalString(pluginOptions.snapshot)
  const envImage = normalizeOptionalString(process.env.DAYTONA_SANDBOX_IMAGE)
  const envSnapshot = normalizeOptionalString(process.env.DAYTONA_SANDBOX_SNAPSHOT)
  const image = envImage ?? configuredImage
  const snapshot = envSnapshot ?? configuredSnapshot

  if (image && snapshot) {
    throw new Error(
      'Configure only one of image or snapshot for the Daytona OpenCode plugin. DAYTONA_SANDBOX_IMAGE and DAYTONA_SANDBOX_SNAPSHOT are mutually exclusive.',
    )
  }

  if (image) {
    return { image }
  }

  if (snapshot) {
    return { snapshot }
  }

  return undefined
}

// Daytona plugin constants

export const LOG_LEVEL_INFO: LogLevel = 'INFO'
export const LOG_LEVEL_ERROR: LogLevel = 'ERROR'
export const LOG_LEVEL_WARN: LogLevel = 'WARN'
