/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString } from 'class-validator'

@ApiSchema({ name: 'PortPreviewUrl' })
export class PortPreviewUrlDto {
  @ApiProperty({
    description: 'ID of the sandbox',
    example: '123456',
  })
  @IsString()
  sandboxId: string

  @ApiProperty({
    description: 'Preview url',
    example: 'https://{port}-{sandboxId}.{proxyDomain',
  })
  @IsString()
  url: string

  @ApiProperty({
    description: 'Access token',
    example: 'ul67qtv-jl6wb9z5o3eii-ljqt9qed6l',
  })
  @IsString()
  token: string
}

@ApiSchema({ name: 'SingedPortPreviewUrl' })
export class SignedPortPreviewUrlDto {
  @ApiProperty({
    description: 'ID of the sandbox',
    example: '123456',
  })
  @IsString()
  sandboxId: string

  @ApiProperty({
    description: 'Singed preview url',
    example: 'https://{port}-{token}.{proxyDomain',
  })
  @IsString()
  url: string
}
