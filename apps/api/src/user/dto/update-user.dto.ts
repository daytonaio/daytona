/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsEnum, IsOptional, IsString } from 'class-validator'
import { SystemRole } from '../enums/system-role.enum'

@ApiSchema({ name: 'UpdateUser' })
export class UpdateUserDto {
  @ApiPropertyOptional()
  @IsString()
  @IsOptional()
  name?: string

  @ApiPropertyOptional()
  @IsString()
  @IsOptional()
  email?: string

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
