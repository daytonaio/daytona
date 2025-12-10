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

  @ApiProperty({
    description: 'SSH command to connect to the sandbox',
    example: 'ssh -p 2222 token@localhost',
  })
  sshCommand: string

  static fromSshAccess(sshAccess: SshAccess, sshGatewayUrl: string): SshAccessDto {
    const dto = new SshAccessDto()
    dto.id = sshAccess.id
    dto.sandboxId = sshAccess.sandboxId
    dto.token = sshAccess.token
    dto.expiresAt = sshAccess.expiresAt
    dto.createdAt = sshAccess.createdAt
    dto.updatedAt = sshAccess.updatedAt
    // Robustly extract host and port from sshGatewayUrl
    let host: string
    let port: string
    try {
      // If protocol is present, use URL
      if (sshGatewayUrl.includes('://')) {
        const url = new URL(sshGatewayUrl)
        host = url.hostname
        port = url.port || '22'
      } else {
        // No protocol, parse manually
        const [hostPart, portPart] = sshGatewayUrl.split(':')
        host = hostPart
        port = portPart || '22'
      }
    } catch {
      // Fallback: treat as host only
      host = sshGatewayUrl
      port = '22'
    }

    if (port === '22') {
      dto.sshCommand = `ssh ${sshAccess.token}@${host}`
    } else {
      dto.sshCommand = `ssh -p ${port} ${sshAccess.token}@${host}`
    }

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

  static fromValidationResult(valid: boolean, sandboxId: string): SshAccessValidationDto {
    const dto = new SshAccessValidationDto()
    dto.valid = valid
    dto.sandboxId = sandboxId
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
