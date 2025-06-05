/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsNumber, IsObject, IsOptional, IsString } from 'class-validator'
import { CreateBuildInfoDto as CreateSnapshotInfoDto } from './create-build-info.dto'

@ApiSchema({ name: 'BuildSnapshot' })
export class BuildSnapshotDto {
  @ApiProperty({
    description: 'The name of the snapshot to build',
    example: 'my-custom-snapshot-v1',
  })
  @IsString()
  @IsNotEmpty()
  name: string

  @ApiPropertyOptional({
    description: 'CPU cores allocated to the resulting sandbox',
    example: 1,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  cpu?: number

  @ApiPropertyOptional({
    description: 'GPU units allocated to the resulting sandbox',
    example: 0,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  gpu?: number

  @ApiPropertyOptional({
    description: 'Memory allocated to the resulting sandbox in GB',
    example: 1,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  memory?: number

  @ApiPropertyOptional({
    description: 'Disk space allocated to the sandbox in GB',
    example: 3,
    type: 'integer',
  })
  @IsOptional()
  @IsNumber()
  disk?: number

  @ApiProperty({
    description: 'Build information for the snapshot',
    type: CreateSnapshotInfoDto,
  })
  @IsObject()
  buildInfo: CreateSnapshotInfoDto
}
