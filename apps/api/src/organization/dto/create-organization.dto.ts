/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsOptional, IsString } from 'class-validator'

@ApiSchema({ name: 'CreateOrganization' })
export class CreateOrganizationDto {
  @ApiProperty({
    description: 'The name of organization',
    example: 'My Organization',
    required: true,
  })
  @IsString()
  @IsNotEmpty()
  name: string

  @ApiPropertyOptional({
    description: 'The region of the organization where region-specific quotas will be applied',
    example: 'us',
    required: false,
    nullable: true,
  })
  @IsString()
  @IsOptional()
  region?: string
}
