/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'
import { ImageState } from '../enums/image-state.enum'
import { Image } from '../entities/image.entity'

export class ImageDto {
  @ApiProperty()
  id: string

  @ApiPropertyOptional()
  organizationId?: string

  @ApiProperty()
  general: boolean

  @ApiProperty()
  name: string

  @ApiProperty()
  enabled: boolean

  @ApiProperty({
    enum: ImageState,
    enumName: 'ImageState',
  })
  state: ImageState

  @ApiProperty({ nullable: true })
  size?: number

  @ApiProperty({ nullable: true })
  entrypoint?: string[]

  @ApiProperty({ nullable: true })
  errorReason?: string

  @ApiProperty()
  createdAt: Date

  @ApiProperty()
  updatedAt: Date

  @ApiProperty({ nullable: true })
  lastUsedAt: Date

  static fromImage(image: Image): ImageDto {
    return {
      id: image.id,
      organizationId: image.organizationId,
      general: image.general,
      name: image.name,
      enabled: image.enabled,
      state: image.state,
      size: image.size,
      entrypoint: image.entrypoint,
      errorReason: image.errorReason,
      createdAt: image.createdAt,
      updatedAt: image.updatedAt,
      lastUsedAt: image.lastUsedAt,
    }
  }
}
