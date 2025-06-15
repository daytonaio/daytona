/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean } from 'class-validator'

@ApiSchema({ name: 'ToggleState' })
export class ToggleStateDto {
  @ApiProperty({
    description: 'Enable or disable the snapshot/tag',
    example: true,
  })
  @IsBoolean()
  enabled: boolean
}
