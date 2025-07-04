/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString } from 'class-validator'

@ApiSchema({ name: 'PortPreviewUrl' })
export class PortPreviewUrlDto {
  @ApiProperty({
    description: 'Preview url',
    example: 'https://123456-mysandbox.runner.com',
  })
  @IsString()
  url: string

  @ApiProperty({
    description: 'Access token',
    example: 'ul67qtv-jl6wb9z5o3eii-ljqt9qed6l',
  })
  @IsString()
  token: string

  @ApiProperty({
    description: 'Legacy preview url using runner domain',
    example: 'https://3000-mysandbox.runner.com',
    required: false,
  })
  @IsString()
  legacyProxyUrl?: string
}
