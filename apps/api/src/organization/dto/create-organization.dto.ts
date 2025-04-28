/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { IsNotEmpty, IsString } from 'class-validator'

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
}
