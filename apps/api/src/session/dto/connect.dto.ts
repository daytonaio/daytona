/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { IsEnum, IsInt, IsOptional, IsString, Min, ValidateNested } from 'class-validator'
import { Type } from 'class-transformer'
import { SessionLanguage } from '../enums/session-language.enum'
import { SessionRefDto } from './code-run.dto'

export class SessionConnectRequestDto {
  @ApiPropertyOptional({ type: SessionRefDto })
  @IsOptional()
  @ValidateNested()
  @Type(() => SessionRefDto)
  context?: SessionRefDto

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  template?: string

  @ApiPropertyOptional({ enum: SessionLanguage })
  @IsOptional()
  @IsEnum(SessionLanguage)
  language?: SessionLanguage

  @ApiPropertyOptional()
  @IsOptional()
  @IsInt()
  @Min(0)
  timeout?: number
}

export class SessionConnectResponseDto {
  @ApiProperty()
  wsUrl: string

  @ApiProperty()
  token: string

  @ApiProperty()
  sessionId: string

  @ApiProperty({
    description: 'When this signed connect URL stops being valid (separate from context idle/absolute TTL).',
  })
  expiresAt: string
}
