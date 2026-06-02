/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'UpdateRegion' })
export class UpdateRegionDto {
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

  @ApiProperty({
    description:
      'Provider-prefixed storage region slug (e.g. "aws-us-east-1") used to pin layered volumes to a specific AWS S3 region. Pass null to clear.',
    example: 'aws-us-east-1',
    nullable: true,
    required: false,
  })
  storageRegion?: string | null
}
