/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsOptional, IsString } from 'class-validator'

@ApiSchema({ name: 'CreateBuildInfo' })
export class CreateBuildInfoDto {
  @ApiProperty({
    description: 'The Dockerfile content used for the build',
    example: 'FROM node:14\nWORKDIR /app\nCOPY . .\nRUN npm install\nCMD ["npm", "start"]',
  })
  @IsString()
  @IsNotEmpty()
  dockerfileContent: string

  @ApiPropertyOptional({
    description: 'The context hashes used for the build',
    type: [String],
    example: ['hash1', 'hash2'],
  })
  @IsString({ each: true })
  @IsOptional()
  contextHashes?: string[]
}
