/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { RegionQuota } from '../entities/region-quota.entity'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'

@ApiSchema({ name: 'RegionQuota' })
export class RegionQuotaDto {
  @ApiProperty()
  organizationId: string

  @ApiProperty()
  regionId: string

  @ApiProperty({ enum: SandboxClass, enumName: 'SandboxClass' })
  sandboxClass: SandboxClass

  @ApiProperty()
  totalCpuQuota: number

  @ApiProperty()
  totalMemoryQuota: number

  @ApiProperty()
  totalDiskQuota: number

  @ApiProperty({ nullable: true })
  maxCpuPerSandbox: number | null

  @ApiProperty({ nullable: true })
  maxMemoryPerSandbox: number | null

  @ApiProperty({ nullable: true })
  maxDiskPerSandbox: number | null

  @ApiProperty({ nullable: true })
  maxDiskPerNonEphemeralSandbox: number | null

  constructor(regionQuota: RegionQuota) {
    this.organizationId = regionQuota.organizationId
    this.regionId = regionQuota.regionId
    this.sandboxClass = regionQuota.sandboxClass
    this.totalCpuQuota = regionQuota.totalCpuQuota
    this.totalMemoryQuota = regionQuota.totalMemoryQuota
    this.totalDiskQuota = regionQuota.totalDiskQuota
    this.maxCpuPerSandbox = regionQuota.maxCpuPerSandbox
    this.maxMemoryPerSandbox = regionQuota.maxMemoryPerSandbox
    this.maxDiskPerSandbox = regionQuota.maxDiskPerSandbox
    this.maxDiskPerNonEphemeralSandbox = regionQuota.maxDiskPerNonEphemeralSandbox
  }
}
