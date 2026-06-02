/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsArray, IsEnum, IsOptional } from 'class-validator'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'
import { GpuType } from '../../sandbox/enums/gpu-type.enum'

@ApiSchema({ name: 'UpdateOrganizationRegionQuota' })
export class UpdateOrganizationRegionQuotaDto {
  @ApiPropertyOptional({ enum: SandboxClass, enumName: 'SandboxClass' })
  @IsOptional()
  @IsEnum(SandboxClass)
  sandboxClass?: SandboxClass

  @ApiProperty({ nullable: true })
  totalCpuQuota?: number

  @ApiProperty({ nullable: true })
  totalMemoryQuota?: number

  @ApiProperty({ nullable: true })
  totalDiskQuota?: number

  @ApiProperty({ nullable: true })
  totalGpuQuota?: number

  @ApiPropertyOptional({ enum: GpuType, enumName: 'GpuType', isArray: true, nullable: true })
  @IsOptional()
  @IsArray()
  @IsEnum(GpuType, { each: true })
  allowedGpuTypes?: GpuType[] | null

  @ApiPropertyOptional({ nullable: true })
  maxCpuPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxMemoryPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxDiskPerSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxDiskPerNonEphemeralSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxCpuPerGpuSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxMemoryPerGpuSandbox?: number | null

  @ApiPropertyOptional({ nullable: true })
  maxDiskPerGpuSandbox?: number | null
}
