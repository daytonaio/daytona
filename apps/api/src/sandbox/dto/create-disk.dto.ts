/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsString, IsInt } from 'class-validator'

export class CreateDiskDto {
  @ApiProperty({
    description: 'Disk name',
    example: 'my-disk',
  })
  @IsString()
  name: string

  @ApiProperty({
    description: 'Disk size in GB',
    example: 100,
  })
  @IsInt()
  size: number
}
