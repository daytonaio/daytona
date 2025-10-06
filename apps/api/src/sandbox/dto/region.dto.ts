/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'Region' })
export class RegionDto {
  @ApiProperty({
    description: 'Region name',
    example: 'us',
  })
  name: string

  constructor(name: string) {
    this.name = name
  }
}
