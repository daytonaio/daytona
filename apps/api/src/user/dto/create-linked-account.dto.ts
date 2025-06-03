/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString } from 'class-validator'

@ApiSchema({ name: 'CreateLinkedAccount' })
export class CreateLinkedAccountDto {
  @ApiProperty({
    description: 'The authentication provider of the secondary account',
  })
  @IsString()
  provider: string

  @ApiProperty({
    description: 'The user ID of the secondary account',
  })
  @IsString()
  userId: string
}
