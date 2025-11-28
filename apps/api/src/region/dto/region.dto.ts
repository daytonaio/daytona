/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { Region } from '../entities/region.entity'

@ApiSchema({ name: 'Region' })
export class RegionDto {
  @ApiProperty({
    description: 'Region ID',
    example: '123456789012',
  })
  id: string

  @ApiProperty({
    description: 'Region name',
    example: 'us-east-1',
  })
  name: string

  @ApiProperty({
    description: 'Organization ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
    nullable: true,
    required: false,
  })
  organizationId: string | null

  @ApiProperty({
    description: 'Creation timestamp',
    example: '2023-01-01T00:00:00.000Z',
  })
  createdAt: string

  @ApiProperty({
    description: 'Last update timestamp',
    example: '2023-01-01T00:00:00.000Z',
  })
  updatedAt: string

  static fromRegion(region: Region): RegionDto {
    return {
      id: region.id,
      name: region.name,
      organizationId: region.organizationId,
      createdAt: region.createdAt?.toISOString(),
      updatedAt: region.updatedAt?.toISOString(),
    }
  }
}
