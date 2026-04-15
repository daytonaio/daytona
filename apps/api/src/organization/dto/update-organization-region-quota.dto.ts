/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'UpdateOrganizationRegionQuota' })
export class UpdateOrganizationRegionQuotaDto {
  @ApiProperty({ nullable: true })
  totalCpuQuota?: number

  @ApiProperty({ nullable: true })
  totalMemoryQuota?: number

  @ApiProperty({ nullable: true })
  totalDiskQuota?: number

  @ApiPropertyOptional({ nullable: true })
  maxCpuPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxMemoryPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxDiskPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxDiskPerNonEphemeralSandbox?: number | null
}
