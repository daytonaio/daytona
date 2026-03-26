/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'

export class RegistryPushAccessDto {
  @ApiProperty({
    description: 'Temporary username for registry authentication',
    example: 'temp-user-123',
  })
  username: string

  @ApiProperty({
    description: 'Temporary secret for registry authentication',
    example: 'eyJhbGciOiJIUzI1NiIs...',
  })
  secret: string

  @ApiProperty({
    description: 'Registry URL',
    example: 'registry.example.com',
  })
  registryUrl: string

  @ApiProperty({
    description: 'Registry ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  registryId: string

  @ApiProperty({
    description: 'Registry project ID',
    example: 'library',
  })
  project: string

  @ApiProperty({
    description: 'Token expiration time in ISO format',
    example: '2023-12-31T23:59:59Z',
  })
  expiresAt: string
}
