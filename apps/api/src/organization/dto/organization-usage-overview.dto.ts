/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'OrganizationUsageOverview' })
export class OrganizationUsageOverviewDto {
  @ApiProperty()
  totalCpuQuota: number
  @ApiProperty()
  totalGpuQuota: number
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
}
