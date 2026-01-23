/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { SandboxState } from '../enums/sandbox-state.enum'
import { WakeOnRequest } from '../enums/wake-on-request.enum'

@ApiSchema({ name: 'SandboxStateInfo' })
export class SandboxStateInfoDto {
  @ApiProperty({
    description: 'The current state of the sandbox',
    enum: SandboxState,
    enumName: 'SandboxState',
    example: SandboxState.STARTED,
  })
  state: SandboxState

  @ApiProperty({
    description: 'Wake on request setting for the sandbox',
    enum: WakeOnRequest,
    enumName: 'WakeOnRequest',
    example: WakeOnRequest.NONE,
  })
  wakeOnRequest: WakeOnRequest
}
