/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional } from '@nestjs/swagger'
import { IsEnum, IsOptional, IsString } from 'class-validator'
import { SessionLanguage } from '../enums/session-language.enum'

export class CreateSessionDto {
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
  @IsString()
  cwd?: string
}
