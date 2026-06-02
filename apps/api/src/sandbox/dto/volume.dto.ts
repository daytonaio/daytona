/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { IsEnum } from 'class-validator'
import { VolumeState } from '../enums/volume-state.enum'
import { Volume } from '../entities/volume.entity'

export class VolumeDto {
  @ApiProperty({
    description: 'Volume ID',
    example: 'vol-12345678',
  })
  id: string

  @ApiProperty({
    description: 'Volume name',
    example: 'my-volume',
  })
  name: string

  @ApiProperty({
    description: 'Organization ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  organizationId: string

  @ApiProperty({
    description: 'Volume state',
    enum: VolumeState,
    enumName: 'VolumeState',
    example: VolumeState.READY,
  })
  @IsEnum(VolumeState)
  state: VolumeState

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

  @ApiPropertyOptional({
    description: 'Last used timestamp',
    example: '2023-01-01T00:00:00.000Z',
    nullable: true,
  })
  lastUsedAt?: string

  @ApiProperty({
    description: 'The error reason of the volume',
    example: 'Error processing volume',
    nullable: true,
  })
  errorReason?: string

  @ApiProperty({
    description:
      'Backend that physically stores the volume. Set when the volume is created from the organization default and immutable afterwards.',
    example: 's3fuse',
    enum: ['s3fuse', 'layered'],
  })
  backend: string

  @ApiPropertyOptional({
    description:
      'Daytona Region ID the volume is pinned to. Populated for layered volumes; null for s3fuse or for legacy layered volumes created before region pinning was introduced.',
    nullable: true,
  })
  regionId?: string | null

  static fromVolume(volume: Volume): VolumeDto {
    return {
      id: volume.id,
      name: volume.name,
      organizationId: volume.organizationId,
      state: volume.state,
      createdAt: volume.createdAt?.toISOString(),
      updatedAt: volume.updatedAt?.toISOString(),
      lastUsedAt: volume.lastUsedAt?.toISOString(),
      errorReason: volume.errorReason,
      backend: volume.backend,
      regionId: volume.regionId ?? null,
    }
  }
}
