/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { AuditLog, AuditLogMetadata } from '../entities/audit-log.entity'

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

  @ApiProperty()
  outcome: string

  @ApiPropertyOptional()
  errorMessage?: string

  @ApiPropertyOptional()
  ipAddress?: string

  @ApiPropertyOptional()
  userAgent?: string

  @ApiPropertyOptional()
  source?: string

  @ApiPropertyOptional()
  metadata?: AuditLogMetadata

  @ApiProperty()
  createdAt: Date

  static fromAuditLog(auditLog: AuditLog): AuditLogDto {
    const dto: AuditLogDto = {
      id: auditLog.id,
      actorId: auditLog.actorId,
      actorEmail: auditLog.actorEmail,
      organizationId: auditLog.organizationId,
      action: auditLog.action,
      targetType: auditLog.targetType,
      targetId: auditLog.targetId,
      outcome: auditLog.outcome,
      errorMessage: auditLog.errorMessage,
      ipAddress: auditLog.ipAddress,
      userAgent: auditLog.userAgent,
      source: auditLog.source,
      metadata: auditLog.metadata,
      createdAt: auditLog.createdAt,
    }

    return dto
  }
}
