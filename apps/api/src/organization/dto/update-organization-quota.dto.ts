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
  maxCpuPerWorkspace?: number

  @ApiProperty({ nullable: true })
  maxMemoryPerWorkspace?: number

  @ApiProperty({ nullable: true })
  maxDiskPerWorkspace?: number

  @ApiProperty({ nullable: true })
  imageQuota?: number

  @ApiProperty({ nullable: true })
  maxImageSize?: number

  @ApiProperty({ nullable: true })
  volumeQuota?: number
}
