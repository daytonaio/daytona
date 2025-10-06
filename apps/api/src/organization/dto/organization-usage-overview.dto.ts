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

  constructor(params: {
    totalCpuQuota: number
    totalMemoryQuota: number
    totalDiskQuota: number
    currentCpuUsage: number
    currentMemoryUsage: number
    currentDiskUsage: number
    totalSnapshotQuota: number
    currentSnapshotUsage: number
    totalVolumeQuota: number
    currentVolumeUsage: number
  }) {
    this.totalCpuQuota = params.totalCpuQuota
    this.totalMemoryQuota = params.totalMemoryQuota
    this.totalDiskQuota = params.totalDiskQuota
    this.currentCpuUsage = params.currentCpuUsage
    this.currentMemoryUsage = params.currentMemoryUsage
    this.currentDiskUsage = params.currentDiskUsage
    this.totalSnapshotQuota = params.totalSnapshotQuota
    this.currentSnapshotUsage = params.currentSnapshotUsage
    this.totalVolumeQuota = params.totalVolumeQuota
    this.currentVolumeUsage = params.currentVolumeUsage
  }
}
