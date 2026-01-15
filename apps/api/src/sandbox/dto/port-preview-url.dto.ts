/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNumber, IsString } from 'class-validator'

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
    example: 'https://{port}-{sandboxId}.{proxyDomain}',
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

@ApiSchema({ name: 'SignedPortPreviewUrl' })
export class SignedPortPreviewUrlDto {
  @ApiProperty({
    description: 'ID of the sandbox',
    example: '123456',
  })
  @IsString()
  sandboxId: string

  @ApiProperty({
    description: 'Port number of the signed preview URL',
    example: 3000,
    type: 'integer',
  })
  @IsNumber()
  port: number

  @ApiProperty({
    description: 'Token of the signed preview URL',
    example: 'jl6wb9z5o3eii',
  })
  @IsString()
  token: string

  @ApiProperty({
    description: 'Signed preview url',
    example: 'https://{port}-{token}.{proxyDomain}',
  })
  @IsString()
  url: string
}
