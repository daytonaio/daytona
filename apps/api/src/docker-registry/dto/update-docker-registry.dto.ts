/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsOptional, IsUrl } from 'class-validator'
import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsSafeDisplayString } from '../../common/validators'

@ApiSchema({ name: 'UpdateDockerRegistry' })
export class UpdateDockerRegistryDto {
  @ApiProperty({ description: 'Registry name' })
  @IsString()
  @IsSafeDisplayString()
  name: string

  @ApiProperty({ description: 'Registry URL' })
  @IsUrl()
  url: string

  @ApiProperty({ description: 'Registry username' })
  @IsString()
  username: string

  @ApiPropertyOptional({ description: 'Registry password' })
  @IsString()
  @IsOptional()
  password?: string

  @ApiPropertyOptional({ description: 'Registry project' })
  @IsString()
  @IsOptional()
  project?: string
}
