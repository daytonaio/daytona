/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'

export class SessionPackageDto {
  @ApiProperty()
  name: string

  @ApiProperty()
  version: string

  @ApiPropertyOptional()
  hasNativeBindings?: boolean
}
