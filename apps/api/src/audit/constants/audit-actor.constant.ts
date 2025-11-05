/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const AuditActorId = {
  SYSTEM: 'system',
} as const

export const AuditSource = {
  SYSTEM: 'system',
  RUNNER: 'runner',
} as const

export type AuditSourceType = (typeof AuditSource)[keyof typeof AuditSource]
