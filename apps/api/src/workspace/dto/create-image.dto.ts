/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsArray, IsBoolean, IsOptional, IsString } from 'class-validator'

@ApiSchema({ name: 'CreateImage' })
export class CreateImageDto {
  @ApiProperty({
    description: 'The name of the image',
    example: 'my-docker-image:8.0.1-alpha',
  })
  @IsString()
  name: string

  @ApiPropertyOptional({
    description: 'The entrypoint command for the image',
    example: 'sleep infinity',
  })
  @IsString({
    each: true,
  })
  @IsArray()
  @IsOptional()
  entrypoint?: string[]

  @ApiPropertyOptional({
    description: 'Whether the image is general',
  })
  @IsBoolean()
  @IsOptional()
  general?: boolean
}
