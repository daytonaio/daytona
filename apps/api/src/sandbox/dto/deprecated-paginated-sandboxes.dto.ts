/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { SandboxDto } from './sandbox.dto'

@ApiSchema({ name: 'DeprecatedPaginatedSandboxes' })
export class DeprecatedPaginatedSandboxesDto {
  @ApiProperty({ type: [SandboxDto] })
  items: SandboxDto[]

  @ApiProperty()
  total: number

  @ApiProperty()
  page: number

  @ApiProperty()
  totalPages: number
}
