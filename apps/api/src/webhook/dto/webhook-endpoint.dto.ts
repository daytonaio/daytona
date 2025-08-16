/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'

export class WebhookEndpointDto {
  @ApiProperty({
    description: 'The unique identifier of the webhook endpoint',
    example: 'end_1234567890abcdef',
  })
  id: string

  @ApiProperty({
    description: 'The URL where webhooks are sent',
    example: 'https://api.example.com/webhooks',
  })
  url: string

  @ApiProperty({
    description: 'A description of the webhook endpoint',
    example: 'Production webhook endpoint',
  })
  description: string

  @ApiProperty({
    description: 'Whether the endpoint is active',
    example: true,
  })
  active: boolean

  @ApiProperty({
    description: 'The event types this endpoint receives',
    example: ['sandbox.created', 'sandbox.updated'],
    type: [String],
  })
  eventTypes: string[]

  @ApiProperty({
    description: 'When the endpoint was created',
    example: '2025-01-01T00:00:00.000Z',
  })
  createdAt: string

  @ApiProperty({
    description: 'When the endpoint was last updated',
    example: '2025-01-01T00:00:00.000Z',
  })
  updatedAt: string
}
