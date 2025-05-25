/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsObject, IsString } from 'class-validator'
import { CreateBuildInfoDto as CreateSnapshotInfoDto } from './create-build-info.dto'

@ApiSchema({ name: 'BuildSnapshot' })
export class BuildSnapshotDto {
  @ApiProperty({
    description: 'The name of the snapshot to build',
    example: 'my-custom-snapshot:1.0.0',
  })
  @IsString()
  @IsNotEmpty()
  name: string

  @ApiProperty({
    description: 'Build information for the snapshot',
    type: CreateSnapshotInfoDto,
  })
  @IsObject()
  buildInfo: CreateSnapshotInfoDto
}
