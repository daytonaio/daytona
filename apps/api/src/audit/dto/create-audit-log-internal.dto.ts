/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditLogMetadata } from '../entities/audit-log.entity'
import { AuditAction } from '../enums/audit-action.enum'
import { AuditTarget } from '../enums/audit-target.enum'

export class CreateAuditLogInternalDto {
  actorId: string
  actorEmail: string
  organizationId?: string
  action: AuditAction
  targetType?: AuditTarget
  targetId?: string
  ipAddress?: string
  userAgent?: string
  source?: string
  statusCode?: number
  errorMessage?: string
  metadata?: AuditLogMetadata
}
