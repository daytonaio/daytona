/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString } from 'class-validator'

@ApiSchema({ name: 'Url' })
export class UrlDto {
  @ApiProperty({
    description: 'URL response',
  })
  @IsString()
  url: string

  constructor(url: string) {
    this.url = url
  }
}
