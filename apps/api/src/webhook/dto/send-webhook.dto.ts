/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsString, IsObject, IsOptional } from 'class-validator'
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'

export class SendWebhookDto {
  @ApiProperty({
    description: 'The type of event being sent',
    example: 'sandbox.created',
  })
  @IsString()
  eventType: string

  @ApiProperty({
    description: 'The payload data to send',
    example: { id: 'sandbox-123', name: 'My Sandbox' },
  })
  @IsObject()
  payload: Record<string, any>

  @ApiPropertyOptional({
    description: 'Optional event ID for idempotency',
    example: 'evt_1234567890abcdef',
  })
  @IsOptional()
  @IsString()
  eventId?: string
}
