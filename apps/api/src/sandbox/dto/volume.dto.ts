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

  constructor(volume: Volume) {
    this.id = volume.id
    this.name = volume.name
    // TODO: Why is org id required in dto
    this.organizationId = volume.organizationId ?? ''
    this.state = volume.state
    this.createdAt = volume.createdAt?.toISOString()
    this.updatedAt = volume.updatedAt?.toISOString()
    this.lastUsedAt = volume.lastUsedAt?.toISOString()
    this.errorReason = volume.errorReason ?? undefined
  }
}
