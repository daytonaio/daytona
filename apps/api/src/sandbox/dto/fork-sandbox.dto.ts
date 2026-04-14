/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsOptional, IsString } from 'class-validator'

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
