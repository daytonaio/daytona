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

  static fromAuditLog(auditLog: AuditLog): AuditLogDto {
    return {
      id: auditLog.id,
      actorId: auditLog.actorId,
      actorEmail: auditLog.actorEmail,
      organizationId: auditLog.organizationId,
      action: auditLog.action,
      targetType: auditLog.targetType,
      targetId: auditLog.targetId,
      statusCode: auditLog.statusCode,
      errorMessage: auditLog.errorMessage,
      ipAddress: auditLog.ipAddress,
      userAgent: auditLog.userAgent,
      source: auditLog.source,
      metadata: auditLog.metadata,
      createdAt: auditLog.createdAt,
    }
  }
}
