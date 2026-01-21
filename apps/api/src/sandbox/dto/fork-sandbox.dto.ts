/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsOptional, IsString } from 'class-validator'
import { SandboxState } from '../enums/sandbox-state.enum'

@ApiSchema({ name: 'ForkSandbox' })
export class ForkSandboxDto {
  @ApiPropertyOptional({
    description: 'The name for the forked sandbox. If not provided, a unique name will be generated.',
    example: 'my-forked-sandbox',
  })
  @IsOptional()
  @IsString()
  name?: string
}

@ApiSchema({ name: 'ForkSandboxResponse' })
export class ForkSandboxResponseDto {
  @ApiProperty({
    description: 'The ID of the newly forked sandbox',
    example: 'forked-sandbox-123',
  })
  id: string

  @ApiProperty({
    description: 'The name of the forked sandbox',
    example: 'my-forked-sandbox',
  })
  name: string

  @ApiProperty({
    description: 'The current state of the forked sandbox',
    enum: SandboxState,
    example: SandboxState.CREATING,
  })
  state: SandboxState

  @ApiProperty({
    description: 'The ID of the parent sandbox that was forked',
    example: 'parent-sandbox-123',
  })
  parentSandboxId: string
}
