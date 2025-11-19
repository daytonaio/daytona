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
    description: 'The ID of the default region for the organization',
    example: 'us',
    required: false,
  })
  @IsString()
  @IsOptional()
  regionId?: string
}
