/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsBoolean } from 'class-validator'

@ApiSchema({ name: 'OrganizationPreviewWarning' })
export class OrganizationPreviewWarningDto {
  @ApiProperty({
    description: 'Whether the proxy shows the preview URL warning page for this organization',
  })
  @IsBoolean()
  previewWarningEnabled: boolean
}
