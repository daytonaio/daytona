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

  static fromSshAccess(sshAccess: SshAccess): SshAccessDto {
    const dto = new SshAccessDto()
    dto.id = sshAccess.id
    dto.sandboxId = sshAccess.sandboxId
    dto.token = sshAccess.token
    dto.expiresAt = sshAccess.expiresAt
    dto.createdAt = sshAccess.createdAt
    dto.updatedAt = sshAccess.updatedAt
    return dto
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

  @ApiProperty({
    description: 'ID of the runner hosting the sandbox',
    example: '123e4567-e89b-12d3-a456-426614174000',
    required: false,
  })
  runnerId?: string

  @ApiProperty({
    description: 'Domain of the runner hosting the sandbox',
    example: 'runner-1.example.com',
    required: false,
  })
  runnerDomain?: string

  static fromValidationResult(
    valid: boolean,
    sandboxId: string,
    runnerId?: string,
    runnerDomain?: string,
  ): SshAccessValidationDto {
    const dto = new SshAccessValidationDto()
    dto.valid = valid
    dto.sandboxId = sandboxId
    dto.runnerId = runnerId
    dto.runnerDomain = runnerDomain
    return dto
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
}
