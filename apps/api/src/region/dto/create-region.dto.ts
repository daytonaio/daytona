/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsString, IsNotEmpty } from 'class-validator'

@ApiSchema({ name: 'CreateRegion' })
export class CreateRegionDto {
  @ApiProperty({
    description: 'Region name',
    example: 'us-east-1',
  })
  @IsString()
  @IsNotEmpty()
  name: string

  @ApiProperty({
    description: 'Proxy URL for the region',
    example: 'https://proxy.example.com',
    nullable: true,
    required: false,
  })
  proxyUrl?: string

  @ApiProperty({
    description: 'SSH Gateway URL for the region',
    example: 'ssh://ssh-gateway.example.com',
    nullable: true,
    required: false,
  })
  sshGatewayUrl?: string

  @ApiProperty({
    description: 'Snapshot Manager URL for the region',
    example: 'https://snapshot-manager.example.com',
    nullable: true,
    required: false,
  })
  snapshotManagerUrl?: string
}

@ApiSchema({ name: 'CreateRegionResponse' })
export class CreateRegionResponseDto {
  @ApiProperty({
    description: 'ID of the created region',
    example: 'region_12345',
  })
  @IsString()
  @IsNotEmpty()
  id: string

  @ApiProperty({
    description: 'Proxy API key for the region',
    example: 'proxy-api-key-xyz',
    nullable: true,
    required: false,
  })
  proxyApiKey?: string

  @ApiProperty({
    description: 'SSH Gateway API key for the region',
    example: 'ssh-gateway-api-key-abc',
    nullable: true,
    required: false,
  })
  sshGatewayApiKey?: string

  @ApiProperty({
    description: 'Snapshot Manager API key for the region',
    example: 'snapshot-manager-api-key-xyz',
    nullable: true,
    required: false,
  })
  snapshotManagerApiKey?: string

  constructor(params: { id: string; proxyApiKey?: string; sshGatewayApiKey?: string; snapshotManagerApiKey?: string }) {
    this.id = params.id
    this.proxyApiKey = params.proxyApiKey
    this.sshGatewayApiKey = params.sshGatewayApiKey
    this.snapshotManagerApiKey = params.snapshotManagerApiKey
  }
}
