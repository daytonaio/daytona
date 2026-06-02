/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsOptional, IsString } from 'class-validator'

@ApiSchema({ name: 'UpdateOrganizationCustomBucket' })
export class UpdateOrganizationCustomBucketDto {
  @ApiProperty({
    description: 'S3 bucket name to use for layered volumes',
    example: 'my-org-volumes',
  })
  @IsString()
  @IsNotEmpty()
  bucketName: string

  @ApiPropertyOptional({
    description:
      'S3-compatible endpoint URL. Required for non-AWS providers (MinIO, R2, GCS, etc.). Omit for native AWS S3.',
    example: 'https://s3.us-east-1.amazonaws.com',
  })
  @IsString()
  @IsOptional()
  endpoint?: string

  @ApiPropertyOptional({
    description: 'AWS region of the bucket (e.g. "us-east-1"). Used when endpoint is omitted (native AWS).',
    example: 'us-east-1',
  })
  @IsString()
  @IsOptional()
  region?: string

  @ApiProperty({
    description: 'Access key ID with read/write permissions on the bucket',
  })
  @IsString()
  @IsNotEmpty()
  accessKeyId: string

  @ApiProperty({
    description: 'Secret access key',
  })
  @IsString()
  @IsNotEmpty()
  secretAccessKey: string
}
