/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'

export class StorageAccessDto {
  @ApiProperty({
    description: 'Access key for storage authentication',
    example: 'temp-user-123',
  })
  accessKey: string

  @ApiProperty({
    description: 'Secret key for storage authentication',
    example: 'abchbGciOiJIUzI1NiIs...',
  })
  secret: string

  @ApiProperty({
    description: 'Session token for storage authentication',
    example: 'eyJhbGciOiJIUzI1NiIs...',
  })
  sessionToken: string

  @ApiProperty({
    description: 'Storage URL',
    example: 'storage.example.com',
  })
  storageUrl: string

  @ApiProperty({
    description: 'Organization ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  organizationId: string

  @ApiProperty({
    description: 'S3 bucket name',
    example: 'daytona',
  })
  bucket: string
}
