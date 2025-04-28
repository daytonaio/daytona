/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'BuildInfo' })
export class BuildInfoDto {
  @ApiProperty({
    description: 'The unique identifier for the build info',
    example: '123e4567e89b12d3a456426614174000',
  })
  imageRef: string

  @ApiPropertyOptional({
    description: 'The Dockerfile content used for the build',
    example: 'FROM node:14\nWORKDIR /app\nCOPY . .\nRUN npm install\nCMD ["npm", "start"]',
  })
  dockerfileContent?: string

  @ApiPropertyOptional({
    description: 'The context hashes used for the build',
    type: [String],
    example: ['hash1', 'hash2'],
  })
  contextHashes?: string[]

  @ApiProperty({
    description: 'The creation timestamp',
  })
  createdAt: Date

  @ApiProperty({
    description: 'The last update timestamp',
  })
  updatedAt: Date
}
