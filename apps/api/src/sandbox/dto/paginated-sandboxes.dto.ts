/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { SandboxDto } from './sandbox.dto'

@ApiSchema({ name: 'PaginatedSandboxes' })
export class PaginatedSandboxesDto {
  @ApiProperty({
    description: 'List of results for the current page',
    type: [SandboxDto],
  })
  items: SandboxDto[]

  @ApiProperty({
    description: 'Cursor for the next page of results',
    type: String,
    nullable: true,
  })
  nextCursor: string | null
}
