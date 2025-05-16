/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'UsageOverview' })
export class OverviewDto {
  @ApiProperty()
  totalCpuQuota: number
  @ApiProperty()
  totalGpuQuota: number
  @ApiProperty()
  totalMemoryQuota: number
  @ApiProperty()
  totalDiskQuota: number
  @ApiProperty()
  totalWorkspaceQuota: number
  @ApiProperty()
  concurrentWorkspaceQuota: number
  @ApiProperty()
  currentCpuUsage: number
  @ApiProperty()
  currentMemoryUsage: number
  @ApiProperty()
  currentDiskUsage: number
  @ApiProperty()
  currentWorkspaces: number
  @ApiProperty()
  concurrentWorkspaces: number
  @ApiProperty()
  currentImageNumber: number
  @ApiProperty()
  imageQuota: number
  @ApiProperty()
  totalImageSizeQuota: number
  @ApiProperty()
  totalImageSizeUsed: number
  @ApiProperty()
  maxVolumes: number
  @ApiProperty()
  usedVolumes: number
}
