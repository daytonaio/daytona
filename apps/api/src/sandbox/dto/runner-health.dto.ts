/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsArray, IsBoolean, IsNumber, IsOptional, IsString, ValidateNested } from 'class-validator'
import { Type } from 'class-transformer'

@ApiSchema({ name: 'RunnerHealthMetrics' })
export class RunnerHealthMetricsDto {
  @ApiProperty({
    description: 'Current CPU load average',
    example: 0.98,
  })
  @IsNumber()
  currentCpuLoadAverage: number

  @ApiProperty({
    description: 'Current CPU usage percentage',
    example: 45.5,
  })
  @IsNumber()
  currentCpuUsagePercentage: number

  @ApiProperty({
    description: 'Current memory usage percentage',
    example: 60.2,
  })
  @IsNumber()
  currentMemoryUsagePercentage: number

  @ApiProperty({
    description: 'Current disk usage percentage',
    example: 35.8,
  })
  @IsNumber()
  currentDiskUsagePercentage: number

  @ApiProperty({
    description: 'Currently allocated CPU cores',
    example: 8,
  })
  @IsNumber()
  currentAllocatedCpu: number

  @ApiProperty({
    description: 'Currently allocated memory in GiB',
    example: 16,
  })
  @IsNumber()
  currentAllocatedMemoryGiB: number

  @ApiProperty({
    description: 'Currently allocated disk in GiB',
    example: 100,
  })
  @IsNumber()
  currentAllocatedDiskGiB: number

  @ApiProperty({
    description: 'Number of snapshots currently stored',
    example: 5,
  })
  @IsNumber()
  currentSnapshotCount: number

  @ApiProperty({
    description: 'Number of started sandboxes',
    example: 10,
  })
  @IsNumber()
  currentStartedSandboxes: number

  @ApiProperty({
    description: 'Total CPU cores on the runner',
    example: 8,
  })
  @IsNumber()
  cpu: number

  @ApiProperty({
    description: 'Total RAM in GiB on the runner',
    example: 16,
  })
  @IsNumber()
  memoryGiB: number

  @ApiProperty({
    description: 'Total disk space in GiB on the runner',
    example: 100,
  })
  @IsNumber()
  diskGiB: number
}

@ApiSchema({ name: 'RunnerServiceHealth' })
export class RunnerServiceHealthDto {
  @ApiProperty({
    description: 'Name of the service being checked',
    example: 'runner',
  })
  @IsString()
  serviceName: string

  @ApiProperty({
    description: 'Whether the service is healthy',
    example: false,
  })
  @IsBoolean()
  healthy: boolean

  @ApiPropertyOptional({
    description: 'Error message if the service is unhealthy',
    example: 'Cannot connect to the runner',
  })
  @IsOptional()
  @IsString()
  error?: string
}

@ApiSchema({ name: 'RunnerHealthcheck' })
export class RunnerHealthcheckDto {
  @ApiPropertyOptional({
    description: 'Runner metrics',
    type: RunnerHealthMetricsDto,
  })
  @IsOptional()
  metrics?: RunnerHealthMetricsDto

  @ApiPropertyOptional({
    description: 'Health status of individual services on the runner',
    type: [RunnerServiceHealthDto],
  })
  @IsOptional()
  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => RunnerServiceHealthDto)
  serviceHealth?: RunnerServiceHealthDto[]

  @ApiPropertyOptional({
    description: 'Runner domain',
    example: 'runner-123.daytona.example.com',
  })
  @IsOptional()
  domain?: string

  @ApiPropertyOptional({
    description: 'Runner proxy URL',
    example: 'http://proxy.daytona.example.com:8080',
  })
  @IsOptional()
  proxyUrl?: string

  @ApiPropertyOptional({
    description: 'Runner API URL',
    example: 'http://api.daytona.example.com:8080',
  })
  @IsOptional()
  apiUrl?: string

  @ApiProperty({
    description: 'Runner app version',
    example: 'v0.0.0-dev',
  })
  @IsString()
  appVersion: string
}
