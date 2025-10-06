/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import type { UserPublicKey } from '../user.entity'

@ApiSchema({ name: 'UserPublicKey' })
export class UserPublicKeyDto {
  @ApiProperty({
    description: 'Public key',
  })
  key: string

  @ApiProperty({
    description: 'Key name',
  })
  name: string

  constructor(publicKey: UserPublicKey) {
    this.key = publicKey.key
    this.name = publicKey.name
  }
}
