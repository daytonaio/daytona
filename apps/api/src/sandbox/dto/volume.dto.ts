/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { VolumeState } from '../enums/volume-state.enum'

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
    example: VolumeState.READY,
  })
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

  @ApiProperty({
    description: 'Last used timestamp',
    example: '2023-01-01T00:00:00.000Z',
  })
  lastUsedAt: string
}
