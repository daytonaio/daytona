/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsString, IsBoolean, IsOptional, IsArray } from 'class-validator'

@ApiSchema({ name: 'CreateSessionRequest' })
export class CreateSessionRequestDto {
  @ApiProperty({
    description: 'The ID of the session',
    example: 'session-123',
  })
  @IsString()
  sessionId: string
}

@ApiSchema({ name: 'SessionExecuteResponse' })
export class SessionExecuteResponseDto {
  @ApiPropertyOptional({
    description: 'The ID of the executed command',
    example: 'cmd-123',
  })
  @IsString()
  @IsOptional()
  cmdId?: string

  @ApiPropertyOptional({
    description: 'The output of the executed command',
    example: 'total 20\ndrwxr-xr-x  4 user group  128 Mar 15 10:30 .',
  })
  @IsString()
  @IsOptional()
  output?: string

  @ApiPropertyOptional({
    description: 'The exit code of the executed command',
    example: 0,
  })
  @IsOptional()
  code?: number
}

@ApiSchema({ name: 'Command' })
export class CommandDto {
  @ApiProperty({
    description: 'The ID of the command',
    example: 'cmd-123',
  })
  @IsString()
  cmdId: string

  @ApiProperty({
    description: 'The command that was executed',
    example: 'ls -la',
  })
  @IsString()
  command: string

  @ApiProperty({
    description: 'The output of the command',
    example: 'total 20\ndrwxr-xr-x  4 user group  128 Mar 15 10:30 .',
  })
  @IsString()
  output: string

  @ApiProperty({
    description: 'The exit code of the command',
    example: 0,
  })
  exitCode: number
}

@ApiSchema({ name: 'Session' })
export class SessionDto {
  @ApiProperty({
    description: 'The ID of the session',
    example: 'session-123',
  })
  @IsString()
  sessionId: string

  @ApiProperty({
    description: 'The list of commands executed in this session',
    type: [CommandDto],
  })
  @IsArray()
  commands: CommandDto[]
}
