/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'
import { GpuType } from '../../sandbox/enums/gpu-type.enum'

@ApiSchema({ name: 'RegionUsageOverview' })
export class RegionUsageOverviewDto {
  @ApiProperty()
  regionId: string

  @ApiProperty({ enum: SandboxClass, enumName: 'SandboxClass' })
  sandboxClass: SandboxClass

  @ApiProperty()
  totalCpuQuota: number
  @ApiProperty()
  currentCpuUsage: number

  @ApiProperty()
  totalMemoryQuota: number
  @ApiProperty()
  currentMemoryUsage: number

  @ApiProperty()
  totalDiskQuota: number
  @ApiProperty()
  currentDiskUsage: number

  @ApiProperty()
  totalGpuQuota: number
  @ApiProperty()
  currentGpuUsage: number

  @ApiPropertyOptional({ enum: GpuType, enumName: 'GpuType', isArray: true })
  allowedGpuTypes?: GpuType[]

  @ApiProperty({ nullable: true })
  maxCpuPerSandbox: number | null

  @ApiProperty({ nullable: true })
  maxMemoryPerSandbox: number | null

  @ApiProperty({ nullable: true })
  maxDiskPerSandbox: number | null

  @ApiProperty({ nullable: true })
  maxDiskPerNonEphemeralSandbox: number | null

  @ApiProperty({ nullable: true })
  maxCpuPerGpuSandbox: number | null

  @ApiProperty({ nullable: true })
  maxMemoryPerGpuSandbox: number | null

  @ApiProperty({ nullable: true })
  maxDiskPerGpuSandbox: number | null
}

@ApiSchema({ name: 'OrganizationUsageOverview' })
export class OrganizationUsageOverviewDto {
  @ApiProperty({
    type: [RegionUsageOverviewDto],
  })
  regionUsage: RegionUsageOverviewDto[]

  // Snapshot usage
  @ApiProperty()
  totalSnapshotQuota: number
  @ApiProperty()
  currentSnapshotUsage: number

  // Volume usage
  @ApiProperty()
  totalVolumeQuota: number
  @ApiProperty()
  currentVolumeUsage: number
}
