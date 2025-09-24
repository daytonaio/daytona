/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { WebhookInitialization } from '../entities/webhook-initialization.entity'

@ApiSchema({ name: 'WebhookInitializationStatus' })
export class WebhookInitializationStatusDto {
  @ApiProperty({
    description: 'Organization ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  organizationId: string

  @ApiProperty({
    description: 'The ID of the Svix application',
    example: 'app_1234567890',
    nullable: true,
  })
  svixApplicationId?: string

  @ApiProperty({
    description: 'The error reason for the last initialization attempt',
    example: 'Failed to create Svix application',
    nullable: true,
  })
  lastError?: string

  @ApiProperty({
    description: 'The number of times the initialization has been attempted',
    example: 3,
  })
  retryCount: number

  @ApiProperty({
    description: 'When the webhook initialization was created',
    example: '2023-01-01T00:00:00.000Z',
  })
  createdAt: string

  @ApiProperty({
    description: 'When the webhook initialization was last updated',
    example: '2023-01-01T00:00:00.000Z',
  })
  updatedAt: string

  constructor(webhookInitialization: WebhookInitialization) {
    this.organizationId = webhookInitialization.organizationId
    this.svixApplicationId = webhookInitialization.svixApplicationId
    this.lastError = webhookInitialization.lastError
    this.retryCount = webhookInitialization.retryCount
    this.createdAt = webhookInitialization.createdAt.toISOString()
    this.updatedAt = webhookInitialization.updatedAt.toISOString()
  }
}
