/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsNumber, IsOptional } from 'class-validator'

@ApiSchema({ name: 'CreateOrganizationQuota' })
export class CreateOrganizationQuotaDto {
  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  totalCpuQuota?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  totalMemoryQuota?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  totalDiskQuota?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  maxCpuPerWorkspace?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  maxMemoryPerWorkspace?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  maxDiskPerWorkspace?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  snapshotQuota?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  maxSnapshotSize?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  volumeQuota?: number
}
