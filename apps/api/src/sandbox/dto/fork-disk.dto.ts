/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsString } from 'class-validator'

export class ForkDiskDto {
  @ApiProperty({
    description: 'Disk ID of the disk forked from the base disk',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  @IsString()
  newDiskName: string
}
