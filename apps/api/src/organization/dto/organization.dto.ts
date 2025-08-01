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

  static fromOrganization(organization: Organization): OrganizationDto {
    const dto: OrganizationDto = {
      id: organization.id,
      name: organization.name,
      createdBy: organization.createdBy,
      personal: organization.personal,
      createdAt: organization.createdAt,
      updatedAt: organization.updatedAt,
      suspended: organization.suspended,
      suspensionReason: organization.suspensionReason,
      suspendedAt: organization.suspendedAt,
      suspendedUntil: organization.suspendedUntil,
      suspensionCleanupGracePeriodHours: organization.suspensionCleanupGracePeriodHours,
      totalCpuQuota: organization.totalCpuQuota,
      totalMemoryQuota: organization.totalMemoryQuota,
      totalDiskQuota: organization.totalDiskQuota,
      maxCpuPerSandbox: organization.maxCpuPerSandbox,
      maxMemoryPerSandbox: organization.maxMemoryPerSandbox,
      maxDiskPerSandbox: organization.maxDiskPerSandbox,
    }

    return dto
  }
}
