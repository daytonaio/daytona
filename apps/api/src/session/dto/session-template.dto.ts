/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'

export class SessionTemplateDto {
  @ApiProperty()
  name: string

  @ApiPropertyOptional()
  description?: string

  @ApiProperty({ type: [String] })
  languages: string[]

  @ApiPropertyOptional({ type: [String] })
  packages?: string[]
}
