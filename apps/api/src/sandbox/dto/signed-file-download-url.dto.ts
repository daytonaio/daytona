/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString } from 'class-validator'

@ApiSchema({ name: 'SignedFileDownloadUrl' })
export class SignedFileDownloadUrlDto {
  @ApiProperty({
    description: 'ID of the sandbox',
    example: '123456',
  })
  @IsString()
  sandboxId: string

  @ApiProperty({
    description: 'File path in the sandbox',
    example: '/home/daytona/report.pdf',
  })
  @IsString()
  path: string

  @ApiProperty({
    description: 'Token of the signed file download URL',
    example: 'fdla7bx3km9qw2np',
  })
  @IsString()
  token: string

  @ApiProperty({
    description: 'Signed file download URL',
    example: 'https://2280-fdla7bx3km9qw2np.proxy.example.com/files/download',
  })
  @IsString()
  url: string

  @ApiProperty({
    description: 'Expiration time of the signed URL (ISO 8601)',
    example: '2026-04-21T12:30:00.000Z',
  })
  @IsString()
  expiresAt: string
}

@ApiSchema({ name: 'SignedFileDownloadInfo' })
export class SignedFileDownloadInfoDto {
  @ApiProperty({
    description: 'ID of the sandbox',
    example: '123456',
  })
  @IsString()
  sandboxId: string

  @ApiProperty({
    description: 'File path in the sandbox',
    example: '/home/daytona/report.pdf',
  })
  @IsString()
  path: string
}
