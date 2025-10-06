/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { SshAccess } from '../entities/ssh-access.entity'

export class SshAccessDto {
  @ApiProperty({
    description: 'Unique identifier for the SSH access',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  id: string

  @ApiProperty({
    description: 'ID of the sandbox this SSH access is for',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  sandboxId: string

  @ApiProperty({
    description: 'SSH access token',
    example: 'abc123def456ghi789jkl012mno345pqr678stu901vwx234yz',
  })
  token: string

  @ApiProperty({
    description: 'When the SSH access expires',
    example: '2025-01-01T12:00:00.000Z',
  })
  expiresAt: Date

  @ApiProperty({
    description: 'When the SSH access was created',
    example: '2025-01-01T11:00:00.000Z',
  })
  createdAt: Date

  @ApiProperty({
    description: 'When the SSH access was last updated',
    example: '2025-01-01T11:00:00.000Z',
  })
  updatedAt: Date

  constructor(sshAccess: SshAccess) {
    this.id = sshAccess.id
    this.sandboxId = sshAccess.sandboxId
    this.token = sshAccess.token
    this.expiresAt = sshAccess.expiresAt
    this.createdAt = sshAccess.createdAt
    this.updatedAt = sshAccess.updatedAt
  }
}

export class SshAccessValidationDto {
  @ApiProperty({
    description: 'Whether the SSH access token is valid',
    example: true,
  })
  valid: boolean

  @ApiProperty({
    description: 'ID of the sandbox this SSH access is for',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  sandboxId: string

  constructor(valid: boolean, sandboxId: string) {
    this.valid = valid
    this.sandboxId = sandboxId
  }
}

export class RevokeSshAccessDto {
  @ApiProperty({
    description: 'ID of the sandbox',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  sandboxId: string

  @ApiProperty({
    description: 'SSH access token to revoke',
    example: 'abc123def456ghi789jkl012mno345pqr678stu901vwx234yz',
  })
  token: string

  private constructor() {
    this.sandboxId = ''
    this.token = ''
  }
}
