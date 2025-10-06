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
  organizationId?: string | null
  action: AuditAction
  targetType?: AuditTarget | null
  targetId?: string | null
  statusCode?: number | null
  errorMessage?: string | null
  ipAddress?: string | null
  userAgent?: string | null
  source?: string | null
  metadata?: AuditLogMetadata | null

  constructor(params: {
    actorId: string
    actorEmail: string
    organizationId?: string | null
    action: AuditAction
    targetType?: AuditTarget | null
    targetId?: string | null
    statusCode?: number | null
    errorMessage?: string | null
    ipAddress?: string | null
    userAgent?: string | null
    source?: string | null
    metadata?: AuditLogMetadata | null
  }) {
    this.actorId = params.actorId
    this.actorEmail = params.actorEmail
    this.organizationId = params.organizationId ?? null
    this.action = params.action
    this.targetType = params.targetType ?? null
    this.targetId = params.targetId ?? null
    this.statusCode = params.statusCode ?? null
    this.errorMessage = params.errorMessage ?? null
    this.ipAddress = params.ipAddress ?? null
    this.userAgent = params.userAgent ?? null
    this.source = params.source ?? null
    this.metadata = params.metadata ?? null
  }
}
