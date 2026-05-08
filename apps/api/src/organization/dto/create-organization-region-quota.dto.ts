/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsEnum, IsNumber, IsOptional } from 'class-validator'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'

@ApiSchema({ name: 'CreateOrganizationRegionQuota' })
export class CreateOrganizationRegionQuotaDto {
  @ApiProperty({ enum: SandboxClass, enumName: 'SandboxClass' })
  @IsEnum(SandboxClass)
  sandboxClass: SandboxClass

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
