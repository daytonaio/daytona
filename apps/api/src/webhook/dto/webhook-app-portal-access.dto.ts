/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'WebhookAppPortalAccess' })
export class WebhookAppPortalAccessDto {
  @ApiProperty({
    description: 'The authentication token for the Svix consumer app portal',
    example: 'appsk_...',
  })
  token: string

  @ApiProperty({
    description: 'The URL to the webhook app portal',
    example: 'https://app.svix.com/login#key=...',
  })
  url: string
}
