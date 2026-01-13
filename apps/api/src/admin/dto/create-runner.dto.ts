/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsNumber, IsOptional, IsString } from 'class-validator'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { CreateRunnerDto } from '../../sandbox/dto/create-runner.dto'

@ApiSchema({ name: 'AdminCreateRunner' })
export class AdminCreateRunnerDto extends CreateRunnerDto {
  @IsString()
  @ApiProperty()
  apiKey: string

  @IsString()
  @ApiProperty({
    description: 'The api version of the runner to create',
    pattern: '^(0|2)$',
    example: '2',
  })
  apiVersion: '0' | '2'

  @ApiProperty({
    required: false,
    description: 'The domain of the runner',
    example: 'runner1.example.com',
  })
  @IsString()
  @IsOptional()
  domain?: string

  @IsString()
  @ApiProperty({
    description: 'The API URL of the runner',
    example: 'https://api.runner1.example.com',
    required: false,
  })
  @IsOptional()
  apiUrl?: string

  @IsString()
  @ApiProperty({
    description: 'The proxy URL of the runner',
    example: 'https://proxy.runner1.example.com',
    required: false,
  })
  @IsOptional()
  proxyUrl?: string

  @IsNumber()
  @ApiProperty({
    description: 'The CPU capacity of the runner',
    example: 8,
    required: false,
  })
  @IsOptional()
  cpu?: number

  @IsNumber()
  @ApiProperty({
    description: 'The memory capacity of the runner in GiB',
    example: 16,
    required: false,
  })
  @IsOptional()
  memoryGiB?: number

  @IsNumber()
  @ApiProperty({
    description: 'The disk capacity of the runner in GiB',
    example: 100,
    required: false,
  })
  @IsOptional()
  diskGiB?: number
}
