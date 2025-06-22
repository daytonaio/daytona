/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsEnum, IsOptional } from 'class-validator'
import { AuditAction } from '../enums/audit-action.enum'
import { AuditTarget } from '../enums/audit-target.enum'

@ApiSchema({ name: 'CreateAuditLog' })
export class CreateAuditLogDto {
  @ApiProperty()
  actorId: string

  @ApiProperty()
  actorEmail: string

  @ApiPropertyOptional()
  @IsOptional()
  organizationId?: string

  @ApiProperty({
    enum: AuditAction,
  })
  @IsEnum(AuditAction)
  action: AuditAction

  @ApiPropertyOptional({
    enum: AuditTarget,
  })
  @IsOptional()
  @IsEnum(AuditTarget)
  targetType?: AuditTarget

  @ApiPropertyOptional()
  @IsOptional()
  targetId?: string
}
