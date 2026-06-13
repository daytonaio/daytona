/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsOptional, IsString } from 'class-validator'

@ApiSchema({ name: 'UpdateRegion' })
export class UpdateRegionDto {
  @ApiProperty({
    description: 'Proxy URL for the region',
    example: 'https://proxy.example.com',
    nullable: true,
    required: false,
  })
  @IsString()
  @IsOptional()
  proxyUrl?: string

  @ApiProperty({
    description: 'SSH Gateway URL for the region',
    example: 'ssh://ssh-gateway.example.com',
    nullable: true,
    required: false,
  })
  @IsString()
  @IsOptional()
  sshGatewayUrl?: string

  @ApiProperty({
    description: 'Snapshot Manager URL for the region',
    example: 'https://snapshot-manager.example.com',
    nullable: true,
    required: false,
  })
  @IsString()
  @IsOptional()
  snapshotManagerUrl?: string
}
