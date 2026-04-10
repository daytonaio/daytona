/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsUrl, IsOptional } from 'class-validator'
import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'CreateDockerRegistry' })
export class CreateDockerRegistryDto {
  @ApiProperty({ description: 'Registry name' })
  @IsString()
  name: string

  @ApiProperty({ description: 'Registry URL' })
  @IsUrl()
  url: string

  @ApiProperty({ description: 'Registry username' })
  @IsString()
  username: string

  @ApiProperty({ description: 'Registry password' })
  @IsString()
  password: string

  @ApiPropertyOptional({ description: 'Registry project' })
  @IsString()
  @IsOptional()
  project?: string
}
