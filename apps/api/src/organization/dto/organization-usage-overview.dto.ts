/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'OrganizationUsageOverview' })
export class OrganizationUsageOverviewDto {
  // Sandbox usage
  @ApiProperty()
  totalCpuQuota: number
  @ApiProperty()
  totalMemoryQuota: number
  @ApiProperty()
  totalDiskQuota: number

  @ApiProperty()
  currentCpuUsage: number
  @ApiProperty()
  currentMemoryUsage: number
  @ApiProperty()
  currentDiskUsage: number

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
