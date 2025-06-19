/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditLogMetadata } from '../entities/audit-log.entity'
import { AuditAction } from '../enums/audit-action.enum'
import { AuditOutcome } from '../enums/audit-outcome-enum'
import { AuditTarget } from '../enums/audit-target.enum'

export class CreateAuditLogInternalDto {
  userId: string
  userEmail: string
  organizationId?: string
  action: AuditAction
  targetType?: AuditTarget
  targetId?: string
  ipAddress?: string
  userAgent?: string
  source?: string
  outcome: AuditOutcome
  errorMessage?: string
  metadata?: AuditLogMetadata
}
