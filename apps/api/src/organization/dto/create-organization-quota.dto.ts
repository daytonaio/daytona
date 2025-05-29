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
  maxCpuPerSandbox?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  maxMemoryPerSandbox?: number

  @ApiPropertyOptional()
  @IsNumber()
  @IsOptional()
  maxDiskPerSandbox?: number

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
