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
    required: false,
  })
  @IsOptional()
  domain?: string

  @ApiProperty({
    description: 'The API URL of the runner',
    example: 'https://api.runner1.example.com',
    required: false,
  })
  @IsOptional()
  apiUrl?: string

  @ApiProperty({
    description: 'The proxy URL of the runner',
    example: 'https://proxy.runner1.example.com',
    required: false,
  })
  @IsOptional()
  proxyUrl?: string

  @ApiProperty({
    description: 'The CPU capacity of the runner',
    example: 8,
  })
  cpu: number

  @ApiProperty({
    description: 'The memory capacity of the runner in GiB',
    example: 16,
  })
  memory: number

  @ApiProperty({
    description: 'The disk capacity of the runner in GiB',
    example: 100,
  })
  disk: number

  @ApiProperty({
    description: 'The GPU capacity of the runner',
    example: 1,
    required: false,
  })
  @IsOptional()
  gpu?: number

  @ApiProperty({
    description: 'The type of GPU',
    required: false,
  })
  @IsOptional()
  gpuType?: string

  @ApiProperty({
    description: 'The class of the runner',
    enum: SandboxClass,
    enumName: 'SandboxClass',
    example: SandboxClass.SMALL,
  })
  @IsEnum(SandboxClass)
  class: SandboxClass

  @ApiPropertyOptional({
    description: 'Current CPU usage percentage',
    example: 45.6,
  })
  currentCpuUsagePercentage: number

  @ApiPropertyOptional({
    description: 'Current RAM usage percentage',
    example: 68.2,
  })
  currentMemoryUsagePercentage: number

  @ApiPropertyOptional({
    description: 'Current disk usage percentage',
    example: 33.8,
  })
  currentDiskUsagePercentage: number

  @ApiPropertyOptional({
    description: 'Current allocated CPU',
    example: 4000,
  })
  currentAllocatedCpu: number

  @ApiPropertyOptional({
    description: 'Current allocated memory in GiB',
    example: 8000,
  })
  currentAllocatedMemoryGiB: number

  @ApiPropertyOptional({
    description: 'Current allocated disk in GiB',
    example: 50000,
  })
  currentAllocatedDiskGiB: number

  @ApiPropertyOptional({
    description: 'Current snapshot count',
    example: 12,
  })
  currentSnapshotCount: number

  @ApiPropertyOptional({
    description: 'Runner availability score',
    example: 85,
  })
  availabilityScore: number

  @ApiProperty({
    description: 'The region of the runner',
    example: 'us',
  })
  region: string

  @ApiProperty({
    description: 'The name of the runner',
    example: 'runner1',
  })
  name: string

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

  @ApiProperty({
    description: 'The version of the runner (deprecated in favor of apiVersion)',
    example: '0',
    deprecated: true,
  })
  version: string

  @ApiProperty({
    description: 'The api version of the runner',
    example: '0',
    deprecated: true,
  })
  apiVersion: string

  @ApiPropertyOptional({
    description: 'The app version of the runner',
    example: 'v0.0.0-dev',
    deprecated: true,
  })
  @IsOptional()
  appVersion?: string

  static fromRunner(runner: Runner): RunnerDto {
    return {
      id: runner.id,
      domain: runner.domain,
      apiUrl: runner.apiUrl,
      proxyUrl: runner.proxyUrl,
      cpu: runner.cpu,
      memory: runner.memoryGiB,
      disk: runner.diskGiB,
      gpu: runner.gpu,
      gpuType: runner.gpuType,
      class: runner.class,
      currentCpuUsagePercentage: runner.currentCpuUsagePercentage,
      currentMemoryUsagePercentage: runner.currentMemoryUsagePercentage,
      currentDiskUsagePercentage: runner.currentDiskUsagePercentage,
      currentAllocatedCpu: runner.currentAllocatedCpu,
      currentAllocatedMemoryGiB: runner.currentAllocatedMemoryGiB,
      currentAllocatedDiskGiB: runner.currentAllocatedDiskGiB,
      currentSnapshotCount: runner.currentSnapshotCount,
      availabilityScore: runner.availabilityScore,
      region: runner.region,
      name: runner.name,
      state: runner.state,
      lastChecked: runner.lastChecked?.toISOString(),
      unschedulable: runner.unschedulable,
      createdAt: runner.createdAt.toISOString(),
      updatedAt: runner.updatedAt.toISOString(),
      version: runner.apiVersion,
      apiVersion: runner.apiVersion,
      appVersion: runner.appVersion,
    }
  }
}
