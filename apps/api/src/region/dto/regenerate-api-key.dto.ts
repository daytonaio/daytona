/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString, IsNotEmpty } from 'class-validator'

@ApiSchema({ name: 'RegenerateApiKeyResponse' })
export class RegenerateApiKeyResponseDto {
  @ApiProperty({
    description: 'The newly generated API key',
    example: 'api-key-xyz123',
  })
  @IsString()
  @IsNotEmpty()
  apiKey: string

  constructor(apiKey: string) {
    this.apiKey = apiKey
  }
}
