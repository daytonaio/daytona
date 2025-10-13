/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
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

  @ApiPropertyOptional({
    description: 'Default region ID',
    required: false,
  })
  defaultRegionId?: string

  @ApiProperty({
    description: 'Authenticated rate limit per minute',
    nullable: true,
  })
  authenticatedRateLimit: number | null

  @ApiProperty({
    description: 'Sandbox create rate limit per minute',
    nullable: true,
  })
  sandboxCreateRateLimit: number | null

  @ApiProperty({
    description: 'Sandbox lifecycle rate limit per minute',
    nullable: true,
  })
  sandboxLifecycleRateLimit: number | null

  @ApiProperty({
    description: 'Experimental configuration',
  })
  experimentalConfig: Record<string, any> | null

  static fromOrganization(organization: Organization): OrganizationDto {
    const experimentalConfig = organization._experimentalConfig
    if (experimentalConfig && experimentalConfig.otel && experimentalConfig.otel.headers) {
      experimentalConfig.otel.headers = Object.entries(experimentalConfig.otel.headers).reduce(
        (acc, [key]) => {
          acc[key] = '******'
          return acc
        },
        {} as Record<string, string>,
      )
    }

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
      maxCpuPerSandbox: organization.maxCpuPerSandbox,
      maxMemoryPerSandbox: organization.maxMemoryPerSandbox,
      maxDiskPerSandbox: organization.maxDiskPerSandbox,
      sandboxLimitedNetworkEgress: organization.sandboxLimitedNetworkEgress,
      defaultRegionId: organization.defaultRegionId,
      authenticatedRateLimit: organization.authenticatedRateLimit,
      sandboxCreateRateLimit: organization.sandboxCreateRateLimit,
      sandboxLifecycleRateLimit: organization.sandboxLifecycleRateLimit,
      experimentalConfig: organization._experimentalConfig,
    }

    return dto
  }
}
