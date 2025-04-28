/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsEnum, IsOptional, IsString } from 'class-validator'
import { SystemRole } from '../enums/system-role.enum'
import { CreateOrganizationQuotaDto } from '../../organization/dto/create-organization-quota.dto'

@ApiSchema({ name: 'CreateUser' })
export class CreateUserDto {
  @ApiProperty()
  @IsString()
  id: string

  @ApiProperty()
  @IsString()
  name: string

  @ApiPropertyOptional()
  @IsString()
  @IsOptional()
  email?: string

  @ApiPropertyOptional()
  @IsOptional()
  personalOrganizationQuota?: CreateOrganizationQuotaDto

  @ApiPropertyOptional({
    enum: SystemRole,
  })
  @IsEnum(SystemRole)
  @IsOptional()
  role?: SystemRole

  @ApiPropertyOptional()
  @IsBoolean()
  @IsOptional()
  emailVerified?: boolean
}
