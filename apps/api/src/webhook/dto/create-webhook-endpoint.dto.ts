/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsUrl, IsOptional, IsArray } from 'class-validator'
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'

export class CreateWebhookEndpointDto {
  @ApiProperty({
    description: 'The URL where webhooks will be sent',
    example: 'https://api.example.com/webhooks',
  })
  @IsUrl()
  url: string

  @ApiPropertyOptional({
    description: 'A description of the webhook endpoint',
    example: 'Production webhook endpoint',
  })
  @IsOptional()
  @IsString()
  description?: string

  @ApiPropertyOptional({
    description: 'Event types to filter webhooks',
    example: ['sandbox.created', 'sandbox.updated'],
    type: [String],
  })
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  eventTypes?: string[]
}
