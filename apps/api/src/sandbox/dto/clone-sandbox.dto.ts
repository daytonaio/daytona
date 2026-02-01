/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsOptional, IsString } from 'class-validator'
import { SandboxState } from '../enums/sandbox-state.enum'

@ApiSchema({ name: 'CloneSandbox' })
export class CloneSandboxDto {
  @ApiPropertyOptional({
    description: 'The name for the cloned sandbox. If not provided, a unique name will be generated.',
    example: 'my-cloned-sandbox',
  })
  @IsOptional()
  @IsString()
  name?: string
}

@ApiSchema({ name: 'CloneSandboxResponse' })
export class CloneSandboxResponseDto {
  @ApiProperty({
    description: 'The ID of the newly cloned sandbox',
    example: 'cloned-sandbox-123',
  })
  id: string

  @ApiProperty({
    description: 'The name of the cloned sandbox',
    example: 'my-cloned-sandbox',
  })
  name: string

  @ApiProperty({
    description: 'The current state of the cloned sandbox',
    enum: SandboxState,
    example: SandboxState.CREATING,
  })
  state: SandboxState

  @ApiProperty({
    description: 'The ID of the source sandbox that was cloned',
    example: 'source-sandbox-123',
  })
  sourceSandboxId: string
}
