/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { Region } from '../entities/region.entity'

export class RegionDto {
  @ApiProperty({
    description: 'Region code',
    example: 'abc12345',
  })
  code: string

  @ApiProperty({
    description: 'Region name',
    example: 'us-east-1',
  })
  name: string

  @ApiProperty({
    description: 'Organization ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  organizationId: string

  @ApiProperty({
    description: 'Docker registry ID (optional)',
    example: '123e4567-e89b-12d3-a456-426614174000',
    nullable: true,
  })
  dockerRegistryId: string | null

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
      code: region.code,
      name: region.name,
      organizationId: region.organizationId,
      dockerRegistryId: region.dockerRegistryId,
      createdAt: region.createdAt?.toISOString(),
      updatedAt: region.updatedAt?.toISOString(),
    }
  }
}
