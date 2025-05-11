/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNumber } from 'class-validator'

@ApiSchema({ name: 'UpdateOrganizationQuota' })
export class UpdateOrganizationQuotaDto {
  @ApiProperty()
  @IsNumber()
  totalCpuQuota: number

  @ApiProperty()
  @IsNumber()
  totalMemoryQuota: number

  @ApiProperty()
  @IsNumber()
  totalDiskQuota: number

  @ApiProperty()
  @IsNumber()
  maxCpuPerWorkspace: number

  @ApiProperty()
  @IsNumber()
  maxMemoryPerWorkspace: number

  @ApiProperty()
  @IsNumber()
  maxDiskPerWorkspace: number

  @ApiProperty()
  @IsNumber()
  maxConcurrentWorkspaces: number

  @ApiProperty()
  @IsNumber()
  workspaceQuota: number

  @ApiProperty()
  @IsNumber()
  imageQuota: number

  @ApiProperty()
  @IsNumber()
  maxImageSize: number

  @ApiProperty()
  @IsNumber()
  totalImageSize: number

  @ApiProperty()
  @IsNumber()
  volumeQuota: number
}
