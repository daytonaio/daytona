/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { AuditLog, type AuditLogMetadata } from '../entities/audit-log.entity'

@ApiSchema({ name: 'AuditLog' })
export class AuditLogDto {
  @ApiProperty()
  id: string

  @ApiProperty()
  actorId: string

  @ApiProperty()
  actorEmail: string

  @ApiPropertyOptional()
  organizationId?: string

  @ApiProperty()
  action: string

  @ApiPropertyOptional()
  targetType?: string

  @ApiPropertyOptional()
  targetId?: string

  @ApiPropertyOptional()
  statusCode?: number

  @ApiPropertyOptional()
  errorMessage?: string

  @ApiPropertyOptional()
  ipAddress?: string

  @ApiPropertyOptional()
  userAgent?: string

  @ApiPropertyOptional()
  source?: string

  @ApiPropertyOptional({
    type: 'object',
    additionalProperties: true,
  })
  metadata?: AuditLogMetadata

  @ApiProperty()
  createdAt: Date

  constructor(auditLog: AuditLog) {
    this.id = auditLog.id
    this.actorId = auditLog.actorId
    this.actorEmail = auditLog.actorEmail
    this.organizationId = auditLog.organizationId ?? undefined
    this.action = auditLog.action
    this.targetType = auditLog.targetType ?? undefined
    this.targetId = auditLog.targetId ?? undefined
    this.statusCode = auditLog.statusCode ?? undefined
    this.errorMessage = auditLog.errorMessage ?? undefined
    this.ipAddress = auditLog.ipAddress ?? undefined
    this.userAgent = auditLog.userAgent ?? undefined
    this.source = auditLog.source ?? undefined
    this.metadata = auditLog.metadata ?? undefined
    this.createdAt = auditLog.createdAt
  }
}
