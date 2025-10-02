/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { SandboxDto } from './sandbox.dto'

@ApiSchema({ name: 'PaginatedSandboxes' })
export class PaginatedSandboxesDto {
  @ApiProperty({ type: [SandboxDto] })
  items: SandboxDto[]

  @ApiProperty()
  total: number

  @ApiProperty()
  page: number

  @ApiProperty()
  totalPages: number
}
