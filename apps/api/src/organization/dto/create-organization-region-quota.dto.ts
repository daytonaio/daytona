/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsNumber, IsOptional } from 'class-validator'

@ApiSchema({ name: 'CreateOrganizationRegionQuota' })
export class CreateOrganizationRegionQuotaDto {
  @ApiProperty()
  @IsNumber()
  totalCpuQuota: number

  @ApiProperty()
  @IsNumber()
  totalMemoryQuota: number

  @ApiProperty()
  @IsNumber()
  totalDiskQuota: number

  @ApiPropertyOptional({ nullable: true })
  @IsNumber()
  @IsOptional()
  maxCpuPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  @IsNumber()
  @IsOptional()
  maxMemoryPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  @IsNumber()
  @IsOptional()
  maxDiskPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  @IsNumber()
  @IsOptional()
  maxDiskPerNonEphemeralSandbox?: number | null
}
