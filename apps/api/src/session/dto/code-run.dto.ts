/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { IsEnum, IsInt, IsObject, IsOptional, IsString, Min, ValidateNested } from 'class-validator'
import { Type } from 'class-transformer'
import { SessionLanguage } from '../enums/session-language.enum'

export class SessionRefDto {
  @ApiProperty({ description: 'The id returned by createSession (or by connect/runStream auto-create).' })
  @IsString()
  id: string
}

export class DisplayDataDto {
  @ApiProperty({ type: [String] })
  formats: string[]

  @ApiProperty({
    type: 'object',
    additionalProperties: { type: 'string' },
    description: 'Mime → payload (base64 for binary).',
  })
  data: Record<string, string>
}

export class ExecutionErrorDto {
  @ApiProperty()
  name: string

  @ApiPropertyOptional()
  value?: string

  @ApiPropertyOptional()
  traceback?: string
}

export class SessionCodeRunRequestDto {
  @ApiProperty()
  @IsString()
  code: string

  @ApiPropertyOptional({ enum: SessionLanguage })
  @IsOptional()
  @IsEnum(SessionLanguage)
  language?: SessionLanguage

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  template?: string

  @ApiPropertyOptional({ type: SessionRefDto })
  @IsOptional()
  @ValidateNested()
  @Type(() => SessionRefDto)
  context?: SessionRefDto

  @ApiPropertyOptional({
    type: 'object',
    additionalProperties: { type: 'string' },
  })
  @IsOptional()
  @IsObject()
  env?: Record<string, string>

  @ApiPropertyOptional({ description: 'Per-call timeout in seconds; 0 means unlimited.' })
  @IsOptional()
  @IsInt()
  @Min(0)
  timeout?: number
}

export class SessionCodeRunResponseDto {
  @ApiProperty()
  stdout: string

  @ApiProperty()
  stderr: string

  @ApiPropertyOptional({ type: ExecutionErrorDto })
  error?: ExecutionErrorDto

  @ApiPropertyOptional({ type: [DisplayDataDto] })
  displays?: DisplayDataDto[]

  @ApiProperty()
  durationMs: number
}
