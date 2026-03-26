/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'RunnerStatus' })
export class RunnerStatusDto {
  @ApiProperty({
    description: 'Current CPU usage percentage',
    example: 45.6,
  })
  currentCpuUsagePercentage: number

  @ApiProperty({
    description: 'Current RAM usage percentage',
    example: 68.2,
  })
  currentMemoryUsagePercentage: number

  @ApiProperty({
    description: 'Current disk usage percentage',
    example: 33.8,
  })
  currentDiskUsagePercentage: number

  @ApiProperty({
    description: 'Current allocated CPU',
    example: 4000,
  })
  currentAllocatedCpu: number

  @ApiProperty({
    description: 'Current allocated memory',
    example: 8000,
  })
  currentAllocatedMemoryGiB: number

  @ApiProperty({
    description: 'Current allocated disk',
    example: 50000,
  })
  currentAllocatedDiskGiB: number

  @ApiProperty({
    description: 'Current snapshot count',
    example: 12,
  })
  currentSnapshotCount: number

  @ApiProperty({
    description: 'Runner status',
    example: 'ok',
  })
  status: string

  @ApiProperty({
    description: 'Runner version',
    example: '0.0.1',
  })
  version: string
}
