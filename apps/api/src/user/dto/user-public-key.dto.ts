/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { UserPublicKey } from '../user.entity'

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

  static fromUserPublicKey(publicKey: UserPublicKey): UserPublicKeyDto {
    const dto: UserPublicKeyDto = {
      key: publicKey.key,
      name: publicKey.name,
    }

    return dto
  }
}
