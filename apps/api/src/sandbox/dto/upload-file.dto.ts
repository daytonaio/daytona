/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'UploadFile' })
export class UploadFileDto {
  @ApiProperty({ type: 'string', format: 'binary' })
  file: any

  @ApiProperty()
  path: string
}
