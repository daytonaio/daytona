/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import { ApiKey } from '../api-key.entity'

@ApiSchema({ name: 'ApiKeyList' })
export class ApiKeyListDto {
  @ApiProperty({
    description: 'The name of the API key',
    example: 'My API Key',
  })
  name: string

  @ApiProperty({
    description: 'The masked API key value',
    example: 'bb_********************def',
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
    description: 'When the API key was last used',
    example: '2024-03-14T12:00:00.000Z',
    nullable: true,
  })
  lastUsedAt?: Date

  @ApiProperty({
    description: 'When the API key expires',
    example: '2024-03-14T12:00:00.000Z',
    nullable: true,
  })
  expiresAt?: Date

  @ApiProperty({
    description: 'The user ID of the user who created the API key',
    example: '123',
  })
  userId: string

  constructor(apiKey: ApiKey) {
    this.name = apiKey.name
    this.value = `${apiKey.keyPrefix}********************${apiKey.keySuffix}`
    this.createdAt = apiKey.createdAt
    this.permissions = apiKey.permissions
    this.lastUsedAt = apiKey.lastUsedAt ?? undefined
    this.expiresAt = apiKey.expiresAt ?? undefined
    this.userId = apiKey.userId
  }
}
