/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsString, IsIn, IsOptional } from 'class-validator'

export class DeviceApproveRequestDto {
  @ApiProperty({ example: 'WDJB-MJHT' })
  @IsString()
  user_code: string

  @ApiProperty({ example: 'approve', enum: ['approve', 'deny'] })
  @IsString()
  @IsIn(['approve', 'deny'])
  action: 'approve' | 'deny'

  @ApiProperty({ example: 'org-abc123', required: false })
  @IsString()
  @IsOptional()
  organization_id?: string
}
