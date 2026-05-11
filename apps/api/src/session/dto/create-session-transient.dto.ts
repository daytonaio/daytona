/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiPropertyOptional } from '@nestjs/swagger'
import { IsEnum, IsOptional, IsString } from 'class-validator'
import { SessionLanguage } from '../enums/session-language.enum'

/**
 * Body for `POST /sessions/transients` — the SDK's one-shot entrypoint. Same
 * shape as `CreateSessionDto` minus `cwd` (transients reuse a single
 * deterministic context per (instance, language) and recycle globals on every
 * exec, so a stable cwd doesn't make sense).
 */
export class CreateSessionTransientDto {
  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  template?: string

  @ApiPropertyOptional({ enum: SessionLanguage })
  @IsOptional()
  @IsEnum(SessionLanguage)
  language?: SessionLanguage
}
