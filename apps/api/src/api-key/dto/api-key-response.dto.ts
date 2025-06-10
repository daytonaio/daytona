/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { ApiKey } from '../api-key.entity'

@ApiSchema({ name: 'ApiKeyResponse' })
export class ApiKeyResponseDto {
  @ApiProperty({
    description: 'The name of the API key',
    example: 'My API Key',
  })
  name: string

  @ApiProperty({
    description: 'The API key value',
    example: 'bb_sk_1234567890abcdef',
  })
  value: string

  @ApiProperty({
    description: 'When the API key was created',
    example: '2024-03-14T12:00:00.000Z',
  })
  createdAt: Date

  @ApiProperty({
    description: 'The list of organization resource permissions assigned to the API key',
    enum: OrganizationResourcePermission,
    isArray: true,
  })
  permissions: OrganizationResourcePermission[]

  @ApiProperty({
    description: 'When the API key expires',
    example: '2025-06-09T12:00:00.000Z',
    nullable: true,
  })
  expiresAt?: Date

  static fromApiKey(apiKey: ApiKey, value: string): ApiKeyResponseDto {
    return {
      name: apiKey.name,
      value,
      createdAt: apiKey.createdAt,
      permissions: apiKey.permissions,
      expiresAt: apiKey.expiresAt,
    }
  }
}
