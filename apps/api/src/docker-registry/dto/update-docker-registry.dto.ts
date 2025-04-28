/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsOptional } from 'class-validator'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'UpdateDockerRegistry' })
export class UpdateDockerRegistryDto {
  @ApiProperty({
    description: 'Registry name',
    required: true,
  })
  @IsString()
  name: string

  @ApiProperty({
    description: 'Registry username',
    required: true,
  })
  @IsString()
  username: string

  @ApiProperty({
    description: 'Registry password',
    required: false,
  })
  @IsString()
  @IsOptional()
  password?: string
}
