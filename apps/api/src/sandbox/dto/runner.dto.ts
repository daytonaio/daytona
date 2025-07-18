/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsEnum, IsOptional } from 'class-validator'
import { Runner } from '../entities/runner.entity'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerState } from '../enums/runner-state.enum'

@ApiSchema({ name: 'Runner' })
export class RunnerDto {
  @ApiProperty({
    description: 'The ID of the runner',
    example: 'runner123',
  })
  id: string

  @ApiProperty({
    description: 'The domain of the runner',
    example: 'runner1.example.com',
  })
  domain: string

  @ApiProperty({
    description: 'The API URL of the runner',
    example: 'https://api.runner1.example.com',
  })
  apiUrl: string

  @ApiProperty({
    description: 'The API key for the runner',
    example: 'api-key-123',
  })
  apiKey: string

  @ApiProperty({
    description: 'The CPU capacity of the runner',
    example: 8,
  })
  cpu: number

  @ApiProperty({
    description: 'The memory capacity of the runner in GB',
    example: 16,
  })
  memory: number

  @ApiProperty({
    description: 'The disk capacity of the runner in GB',
    example: 100,
  })
  disk: number

  @ApiProperty({
    description: 'The GPU capacity of the runner',
    example: 1,
  })
  gpu: number

  @ApiProperty({
    description: 'The type of GPU',
  })
  gpuType: string

  @ApiProperty({
    description: 'The class of the runner',
    enum: SandboxClass,
    enumName: 'SandboxClass',
    example: SandboxClass.SMALL,
  })
  @IsEnum(SandboxClass)
  class: SandboxClass

  @ApiProperty({
    description: 'The current usage of the runner',
    example: 2,
  })
  used: number

  @ApiProperty({
    description: 'The capacity of the runner',
    example: 10,
  })
  capacity: number

  @ApiProperty({
    description: 'The region of the runner',
    example: 'us',
  })
  region: string

  @ApiProperty({
    description: 'The state of the runner',
    enum: RunnerState,
    enumName: 'RunnerState',
    example: RunnerState.INITIALIZING,
  })
  @IsEnum(RunnerState)
  state: RunnerState

  @ApiPropertyOptional({
    description: 'The last time the runner was checked',
    example: '2024-10-01T12:00:00Z',
    required: false,
  })
  @IsOptional()
  lastChecked?: string

  @ApiProperty({
    description: 'Whether the runner is unschedulable',
    example: false,
  })
  unschedulable: boolean

  @ApiProperty({
    description: 'The creation timestamp of the runner',
    example: '2023-10-01T12:00:00Z',
  })
  createdAt: string

  @ApiProperty({
    description: 'The last update timestamp of the runner',
    example: '2023-10-01T12:00:00Z',
  })
  updatedAt: string

  static fromRunner(runner: Runner): RunnerDto {
    return {
      id: runner.id,
      domain: runner.domain,
      apiUrl: runner.apiUrl,
      apiKey: runner.apiKey,
      cpu: runner.cpu,
      memory: runner.memory,
      disk: runner.disk,
      gpu: runner.gpu,
      gpuType: runner.gpuType,
      class: runner.class,
      used: runner.used,
      capacity: runner.capacity,
      region: runner.region,
      state: runner.state,
      lastChecked: runner.lastChecked?.toISOString(),
      unschedulable: runner.unschedulable,
      createdAt: runner.createdAt.toISOString(),
      updatedAt: runner.updatedAt.toISOString(),
    }
  }
}
