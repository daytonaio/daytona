/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { User } from '../user.entity'
import { UserPublicKeyDto } from './user-public-key.dto'

@ApiSchema({ name: 'User' })
export class UserDto {
  @ApiProperty({
    description: 'User ID',
  })
  id: string

  @ApiProperty({
    description: 'User name',
  })
  name: string

  @ApiProperty({
    description: 'User email',
  })
  email: string

  @ApiProperty({
    description: 'User public keys',
    type: [UserPublicKeyDto],
  })
  publicKeys: UserPublicKeyDto[]

  @ApiProperty({
    description: 'Creation timestamp',
  })
  createdAt: Date

  constructor(user: User) {
    this.id = user.id
    this.name = user.name
    this.email = user.email
    this.publicKeys = user.publicKeys.map((publicKey) => new UserPublicKeyDto(publicKey))
    this.createdAt = user.createdAt
  }
}
