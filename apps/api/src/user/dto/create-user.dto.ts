/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsEmail, IsEnum, IsOptional, IsString } from 'class-validator'
import { SystemRole } from '../enums/system-role.enum'
import { CreateOrganizationQuotaDto } from '../../organization/dto/create-organization-quota.dto'
import { IsSafeDisplayString } from '../../common/validators'

@ApiSchema({ name: 'CreateUser' })
export class CreateUserDto {
  @ApiProperty()
  @IsString()
  id: string

  @ApiProperty()
  @IsString()
  @IsSafeDisplayString()
  name: string

  @ApiPropertyOptional()
  @IsEmail()
  @IsOptional()
  email?: string

  @ApiPropertyOptional()
  @IsOptional()
  personalOrganizationQuota?: CreateOrganizationQuotaDto

  @ApiPropertyOptional()
  @IsString()
  @IsOptional()
  personalOrganizationDefaultRegionId?: string

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
