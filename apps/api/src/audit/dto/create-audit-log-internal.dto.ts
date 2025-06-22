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
  statusCode?: number
  errorMessage?: string
  ipAddress?: string
  userAgent?: string
  source?: string
  metadata?: AuditLogMetadata
}
