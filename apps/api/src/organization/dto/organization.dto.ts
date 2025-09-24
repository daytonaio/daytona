/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { Organization } from '../entities/organization.entity'

@ApiSchema({ name: 'Organization' })
export class OrganizationDto {
  @ApiProperty({
    description: 'Organization ID',
  })
  id: string

  @ApiProperty({
    description: 'Organization name',
  })
  name: string

  @ApiProperty({
    description: 'User ID of the organization creator',
  })
  createdBy: string

  @ApiProperty({
    description: 'Personal organization flag',
  })
  personal: boolean

  @ApiProperty({
    description: 'Creation timestamp',
  })
  createdAt: Date

  @ApiProperty({
    description: 'Last update timestamp',
  })
  updatedAt: Date

  @ApiProperty({
    description: 'Suspended flag',
  })
  suspended: boolean

  @ApiProperty({
    description: 'Suspended at',
  })
  suspendedAt?: Date

  @ApiProperty({
    description: 'Suspended reason',
  })
  suspensionReason?: string

  @ApiProperty({
    description: 'Suspended until',
  })
  suspendedUntil?: Date

  @ApiProperty({
    description: 'Suspension cleanup grace period hours',
  })
  suspensionCleanupGracePeriodHours?: number

  @ApiProperty({
    description: 'Total CPU quota',
  })
  totalCpuQuota: number

  @ApiProperty({
    description: 'Total memory quota',
  })
  totalMemoryQuota: number

  @ApiProperty({
    description: 'Total disk quota',
  })
  totalDiskQuota: number

  @ApiProperty({
    description: 'Max CPU per sandbox',
  })
  maxCpuPerSandbox: number

  @ApiProperty({
    description: 'Max memory per sandbox',
  })
  maxMemoryPerSandbox: number

  @ApiProperty({
    description: 'Max disk per sandbox',
  })
  maxDiskPerSandbox: number

  @ApiProperty({
    description: 'Sandbox default network block all',
  })
  sandboxLimitedNetworkEgress: boolean

  constructor(organization: Organization) {
    this.id = organization.id
    this.name = organization.name
    this.createdBy = organization.createdBy
    this.personal = organization.personal
    this.createdAt = organization.createdAt
    this.updatedAt = organization.updatedAt
    this.suspended = organization.suspended
    this.suspensionReason = organization.suspensionReason ?? undefined
    this.suspendedAt = organization.suspendedAt ?? undefined
    this.suspendedUntil = organization.suspendedUntil ?? undefined
    this.suspensionCleanupGracePeriodHours = organization.suspensionCleanupGracePeriodHours
    this.totalCpuQuota = organization.totalCpuQuota
    this.totalMemoryQuota = organization.totalMemoryQuota
    this.totalDiskQuota = organization.totalDiskQuota
    this.maxCpuPerSandbox = organization.maxCpuPerSandbox
    this.maxMemoryPerSandbox = organization.maxMemoryPerSandbox
    this.maxDiskPerSandbox = organization.maxDiskPerSandbox
    this.sandboxLimitedNetworkEgress = organization.sandboxLimitedNetworkEgress
  }
}
