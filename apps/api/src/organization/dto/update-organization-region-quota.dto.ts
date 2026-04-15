/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'UpdateOrganizationRegionQuota' })
export class UpdateOrganizationRegionQuotaDto {
  @ApiProperty({ nullable: true })
  totalCpuQuota?: number

  @ApiProperty({ nullable: true })
  totalMemoryQuota?: number

  @ApiProperty({ nullable: true })
  totalDiskQuota?: number

  @ApiProperty({ nullable: true })
  maxCpuPerSandbox?: number | null

  @ApiProperty({ nullable: true })
  maxMemoryPerSandbox?: number | null

  @ApiProperty({ nullable: true })
  maxDiskPerSandbox?: number | null
}
