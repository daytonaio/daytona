/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsObject, IsString } from 'class-validator'
import { CreateBuildInfoDto } from './create-build-info.dto'

@ApiSchema({ name: 'BuildImage' })
export class BuildImageDto {
  @ApiProperty({
    description: 'The name of the image to build',
    example: 'my-custom-image:1.0.0',
  })
  @IsString()
  @IsNotEmpty()
  name: string

  @ApiProperty({
    description: 'Build information for the image',
    type: CreateBuildInfoDto,
  })
  @IsObject()
  buildInfo: CreateBuildInfoDto
}
