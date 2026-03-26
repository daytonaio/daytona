/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsEnum } from 'class-validator'
import { Region } from '../entities/region.entity'
import { RegionType } from '../enums/region-type.enum'

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
    description: 'The type of the region',
    enum: RegionType,
    enumName: 'RegionType',
    example: Object.values(RegionType)[0],
  })
  @IsEnum(RegionType)
  regionType: RegionType

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

  @ApiProperty({
    description: 'Proxy URL for the region',
    example: 'https://proxy.example.com',
    nullable: true,
    required: false,
  })
  proxyUrl?: string | null

  @ApiProperty({
    description: 'SSH Gateway URL for the region',
    example: 'http://ssh-gateway.example.com',
    nullable: true,
    required: false,
  })
  sshGatewayUrl?: string | null

  @ApiProperty({
    description: 'Snapshot Manager URL for the region',
    example: 'http://snapshot-manager.example.com',
    nullable: true,
    required: false,
  })
  snapshotManagerUrl?: string | null

  static fromRegion(region: Region): RegionDto {
    return {
      id: region.id,
      name: region.name,
      organizationId: region.organizationId,
      regionType: region.regionType,
      createdAt: region.createdAt?.toISOString(),
      updatedAt: region.updatedAt?.toISOString(),
      proxyUrl: region.proxyUrl,
      sshGatewayUrl: region.sshGatewayUrl,
      snapshotManagerUrl: region.snapshotManagerUrl,
    }
  }
}
