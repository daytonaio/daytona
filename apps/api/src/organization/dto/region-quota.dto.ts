/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { RegionQuota } from '../entities/region-quota.entity'

@ApiSchema({ name: 'RegionQuota' })
export class RegionQuotaDto {
  @ApiProperty()
  organizationId: string

  @ApiProperty()
  regionId: string

  @ApiProperty()
  totalCpuQuota: number

  @ApiProperty()
  totalMemoryQuota: number

  @ApiProperty()
  totalDiskQuota: number

  @ApiProperty()
  totalGpuQuota: number

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

  constructor(regionQuota: RegionQuota) {
    this.organizationId = regionQuota.organizationId
    this.regionId = regionQuota.regionId
    this.totalCpuQuota = regionQuota.totalCpuQuota
    this.totalMemoryQuota = regionQuota.totalMemoryQuota
    this.totalDiskQuota = regionQuota.totalDiskQuota
    this.totalGpuQuota = regionQuota.totalGpuQuota
    this.maxCpuPerSandbox = regionQuota.maxCpuPerSandbox
    this.maxMemoryPerSandbox = regionQuota.maxMemoryPerSandbox
    this.maxDiskPerSandbox = regionQuota.maxDiskPerSandbox
    this.maxDiskPerNonEphemeralSandbox = regionQuota.maxDiskPerNonEphemeralSandbox
    this.maxCpuPerGpuSandbox = regionQuota.maxCpuPerGpuSandbox
    this.maxMemoryPerGpuSandbox = regionQuota.maxMemoryPerGpuSandbox
    this.maxDiskPerGpuSandbox = regionQuota.maxDiskPerGpuSandbox
  }
}
