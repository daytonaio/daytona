/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { Runner } from '../entities/runner.entity'

@ApiSchema({ name: 'CreateRunnerResponse' })
export class CreateRunnerResponseDto {
  @ApiProperty({
    description: 'The ID of the runner',
    example: 'runner123',
  })
  id: string

  @ApiProperty({
    description: 'The API key for the runner',
    example: 'dtn_1234567890',
  })
  apiKey: string

  static fromRunner(runner: Runner, apiKey: string): CreateRunnerResponseDto {
    return {
      id: runner.id,
      apiKey,
    }
  }
}
