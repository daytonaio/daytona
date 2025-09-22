/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsEnum } from 'class-validator'
import { SandboxState } from '../enums/sandbox-state.enum'

export class UpdateSandboxStateDto {
  @IsEnum(SandboxState)
  @ApiProperty({
    description: 'The new state for the sandbox',
    enum: SandboxState,
    example: SandboxState.STARTED,
  })
  state: SandboxState
}
