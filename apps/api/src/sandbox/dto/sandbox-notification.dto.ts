/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'

export class SandboxNotificationDto {
  @ApiProperty({
    description: 'The notification message',
    example: 'Your sandbox is running low on memory',
  })
  message: string
}
