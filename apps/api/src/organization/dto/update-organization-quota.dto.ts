/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNumber } from 'class-validator'

@ApiSchema({ name: 'UpdateOrganizationQuota' })
export class UpdateOrganizationQuotaDto {
  @ApiProperty({ nullable: true })
  totalCpuQuota?: number

  @ApiProperty({ nullable: true })
  totalMemoryQuota?: number

  @ApiProperty({ nullable: true })
  totalDiskQuota?: number

  @ApiProperty({ nullable: true })
  maxCpuPerSandbox?: number

  @ApiProperty({ nullable: true })
  maxMemoryPerSandbox?: number

  @ApiProperty({ nullable: true })
  maxDiskPerSandbox?: number

  @ApiProperty({ nullable: true })
  snapshotQuota?: number

  @ApiProperty({ nullable: true })
  maxSnapshotSize?: number

  @ApiProperty({ nullable: true })
  volumeQuota?: number
}
