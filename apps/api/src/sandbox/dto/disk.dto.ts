/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { IsEnum } from 'class-validator'
import { DiskState } from '../enums/disk-state.enum'
import { Disk } from '../entities/disk.entity'

export class DiskDto {
  @ApiProperty({
    description: 'Disk ID',
    example: 'disk-12345678',
  })
  id: string

  @ApiProperty({
    description: 'Organization ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  organizationId: string

  @ApiProperty({
    description: 'Disk name',
    example: 'my-disk',
  })
  name: string

  @ApiProperty({
    description: 'Disk size in GB',
    example: 100,
  })
  size: number

  @ApiProperty({
    description: 'Disk state',
    enum: DiskState,
    enumName: 'DiskState',
    example: DiskState.READY,
  })
  @IsEnum(DiskState)
  state: DiskState

  @ApiPropertyOptional({
    description: 'Runner ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
    nullable: true,
  })
  runnerId?: string

  @ApiPropertyOptional({
    description: 'Sandbox ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
    nullable: true,
  })
  sandboxId?: string

  @ApiPropertyOptional({
    description: 'Error reason',
    example: 'Error processing disk',
    nullable: true,
  })
  errorReason?: string

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

  static fromDisk(disk: Disk): DiskDto {
    return {
      id: disk.id,
      organizationId: disk.organizationId,
      name: disk.name,
      size: disk.size,
      state: disk.state,
      runnerId: disk.runnerId,
      sandboxId: disk.sandboxId,
      errorReason: disk.errorReason,
      createdAt: disk.createdAt?.toISOString(),
      updatedAt: disk.updatedAt?.toISOString(),
    }
  }
}
