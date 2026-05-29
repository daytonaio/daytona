/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { SandboxState } from '../enums/sandbox-state.enum'

@ApiSchema({ name: 'SandboxStateCount' })
export class SandboxStateCountDto {
  @ApiProperty({
    description: 'Sandbox state.',
    enum: SandboxState,
    enumName: 'SandboxState',
    example: SandboxState.STARTED,
  })
  state: SandboxState

  @ApiProperty({
    description: 'Number of sandboxes in this state matching the query filters.',
    example: 12,
  })
  count: number
}

@ApiSchema({ name: 'SandboxesSummary' })
export class SandboxesSummaryDto {
  @ApiProperty({
    description: 'Total number of sandboxes matching the query filters.',
    example: 42,
  })
  total: number

  @ApiProperty({
    description: 'Summary of sandboxes by state.',
    type: [SandboxStateCountDto],
  })
  byState: SandboxStateCountDto[]

  @ApiProperty({
    description: 'Number of sandboxes currently in an errored state that are flagged as recoverable.',
    example: 2,
  })
  recoverableErrorCount: number
}
