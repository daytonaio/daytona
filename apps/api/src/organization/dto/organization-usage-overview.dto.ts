/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'SandboxUsageOverview' })
export class SandboxUsageOverviewDto {
  @ApiProperty()
  region: string

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
}

@ApiSchema({ name: 'OrganizationUsageOverview' })
export class OrganizationUsageOverviewDto {
  @ApiProperty({
    type: [SandboxUsageOverviewDto],
  })
  sandboxUsage: SandboxUsageOverviewDto[]

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
