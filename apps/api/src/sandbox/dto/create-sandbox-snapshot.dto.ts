/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString } from 'class-validator'

@ApiSchema({ name: 'CreateSandboxSnapshot' })
export class CreateSandboxSnapshotDto {
  @ApiProperty({
    description: 'Name for the new snapshot',
    example: 'my-dev-env-v1',
  })
  @IsString()
  name: string
}
