/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsBoolean, IsEmail, IsEnum, IsOptional, IsString } from 'class-validator'
import { SystemRole } from '../enums/system-role.enum'
import { IsSafeDisplayString } from '../../common/validators'

@ApiSchema({ name: 'UpdateUser' })
export class UpdateUserDto {
  @ApiPropertyOptional()
  @IsString()
  @IsOptional()
  @IsSafeDisplayString()
  name?: string

  @ApiPropertyOptional()
  @IsEmail()
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
