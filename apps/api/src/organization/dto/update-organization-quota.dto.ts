/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNumber, IsOptional } from 'class-validator'

@ApiSchema({ name: 'UpdateOrganizationQuota' })
export class UpdateOrganizationQuotaDto {
  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  maxCpuPerSandbox?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  maxMemoryPerSandbox?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  maxDiskPerSandbox?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  snapshotQuota?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  maxSnapshotSize?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  volumeQuota?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  authenticatedRateLimit?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  sandboxCreateRateLimit?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  sandboxLifecycleRateLimit?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  authenticatedRateLimitTtlSeconds?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  sandboxCreateRateLimitTtlSeconds?: number

  @ApiProperty({ nullable: true })
  @IsNumber()
  @IsOptional()
  sandboxLifecycleRateLimitTtlSeconds?: number

  @ApiProperty({ nullable: true, description: 'Time in minutes before an unused snapshot is deactivated' })
  @IsNumber()
  @IsOptional()
  snapshotDeactivationTimeoutMinutes?: number
}
