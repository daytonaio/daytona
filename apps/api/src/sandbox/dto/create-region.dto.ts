/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString, IsNotEmpty, IsOptional } from 'class-validator'

@ApiSchema({ name: 'CreateRegion' })
export class CreateRegionDto {
  @ApiProperty({
    description: 'Region name',
    example: 'us-east-1',
  })
  @IsString()
  @IsNotEmpty()
  name: string

  @ApiProperty({
    description: 'Docker registry ID (optional)',
    example: '123e4567-e89b-12d3-a456-426614174000',
    required: false,
    nullable: true,
  })
  @IsOptional()
  @IsString()
  dockerRegistryId?: string
}
