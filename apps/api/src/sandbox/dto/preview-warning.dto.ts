/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean } from 'class-validator'

@ApiSchema({ name: 'PreviewWarning' })
export class PreviewWarningDto {
  @ApiProperty({
    description: 'Whether the preview warning page is enabled for the sandbox',
  })
  @IsBoolean()
  enabled: boolean
}
