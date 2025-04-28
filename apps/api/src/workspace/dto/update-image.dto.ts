/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean } from 'class-validator'

@ApiSchema({ name: 'SetImageGeneralStatus' })
export class SetImageGeneralStatusDto {
  @ApiProperty({
    description: 'Whether the image is general',
    example: true,
  })
  @IsBoolean()
  general: boolean
}
