/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsArray, IsObject, IsBoolean, IsNumber, IsOptional, IsString } from 'class-validator'
import { CreateBuildInfoDto } from './create-build-info.dto'

@ApiSchema({ name: 'CreateSnapshot' })
export class CreateSnapshotDto {
  @ApiProperty({
    description: 'The name of the snapshot',
    example: 'ubuntu-4vcpu-8ram-100gb',
  })
  @IsString()
  name: string

  @ApiPropertyOptional({
    description: 'The image name of the snapshot',
    example: 'ubuntu:22.04',
  })
  @IsOptional()
  @IsString()
  imageName?: string

  @ApiPropertyOptional({
    description: 'The entrypoint command for the snapshot',
    example: 'sleep infinity',
  })
  @IsString({
    each: true,
  })
  @IsArray()
  @IsOptional()
  entrypoint?: string[]

  @ApiPropertyOptional({
    description: 'Whether the snapshot is general',
  })
  @IsBoolean()
  @IsOptional()
  general?: boolean

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

  @ApiPropertyOptional({
    description: 'Build information for the snapshot',
    type: CreateBuildInfoDto,
  })
  @IsOptional()
  @IsObject()
  buildInfo?: CreateBuildInfoDto
}
