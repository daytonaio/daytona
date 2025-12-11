/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsEnum, IsObject, IsOptional, IsString } from 'class-validator'
import { JobType } from '../enums/job-type.enum'
import { JobStatus } from '../enums/job-status.enum'
import { ResourceType } from '../enums/resource-type.enum'

// Re-export enums for convenience
export { JobType, JobStatus, ResourceType }

// Typed payload interfaces for different job types
// Note: resourceId (sandboxId, snapshotRef, etc.) is stored in Job.resourceId column, not in payload
export interface SandboxJobPayload {
  snapshot?: string
  cpu?: number
  mem?: number
  disk?: number
  metadata?: Record<string, any>
}

export interface SnapshotJobPayload {
  // Source registry for pulling external images
  registry?: {
    url?: string
    username?: string
    password?: string
    project?: string
  }
  // Destination registry for pushing to internal registry
  destinationRegistry?: {
    url?: string
    username?: string
    password?: string
    project?: string
  }
  // Source image name for PULL_SNAPSHOT (external image like ubuntu:22.04)
  sourceImage?: string
  // Destination ref for PULL_SNAPSHOT (internal registry ref)
  destinationRef?: string
  // Build info for BUILD_SNAPSHOT
  buildInfo?: {
    snapshotRef: string
    context?: string
    dockerfile?: string
    buildArgs?: Record<string, string>
  }
  organizationId?: string
}

export interface BackupJobPayload {
  backupSnapshotName: string
  registry?: {
    url?: string
    username?: string
    password?: string
    project?: string
  }
}

@ApiSchema({ name: 'Job' })
export class JobDto {
  @ApiProperty({
    description: 'The ID of the job',
    example: 'job123',
  })
  id: string

  @ApiProperty({
    description: 'The type of the job',
    enum: JobType,
    example: JobType.CREATE_SANDBOX,
  })
  @IsEnum(JobType)
  type: JobType

  @ApiProperty({
    description: 'The status of the job',
    enum: JobStatus,
    example: JobStatus.PENDING,
  })
  @IsEnum(JobStatus)
  status: JobStatus

  @ApiPropertyOptional({
    description: 'The type of resource this job operates on',
    enum: ResourceType,
    example: ResourceType.SANDBOX,
  })
  @IsOptional()
  @IsEnum(ResourceType)
  resourceType?: ResourceType

  @ApiPropertyOptional({
    description: 'The ID of the resource this job operates on (sandboxId, snapshotRef, etc.)',
    example: 'sandbox123',
  })
  @IsOptional()
  @IsString()
  resourceId?: string

  @ApiPropertyOptional({
    description: 'Job-specific payload data (operational metadata)',
    type: 'object',
    additionalProperties: true,
  })
  @IsOptional()
  @IsObject()
  payload?: Record<string, any>

  @ApiPropertyOptional({
    description: 'OpenTelemetry trace context for distributed tracing (W3C Trace Context format)',
    type: 'object',
    additionalProperties: true,
    example: { traceparent: '00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01' },
  })
  @IsOptional()
  @IsObject()
  traceContext?: Record<string, string>

  @ApiPropertyOptional({
    description: 'Error message if the job failed',
    example: 'Failed to create sandbox',
  })
  @IsOptional()
  @IsString()
  errorMessage?: string

  @ApiProperty({
    description: 'The creation timestamp of the job',
    example: '2024-10-01T12:00:00Z',
  })
  createdAt: string

  @ApiPropertyOptional({
    description: 'The last update timestamp of the job',
    example: '2024-10-01T12:00:00Z',
  })
  @IsOptional()
  updatedAt?: string
}

@ApiSchema({ name: 'PollJobsRequest' })
export class PollJobsRequestDto {
  @ApiPropertyOptional({
    description: 'Timeout in seconds for long polling (default: 30)',
    example: 30,
  })
  @IsOptional()
  timeout?: number

  @ApiPropertyOptional({
    description: 'Maximum number of jobs to return',
    example: 10,
  })
  @IsOptional()
  limit?: number
}

@ApiSchema({ name: 'PollJobsResponse' })
export class PollJobsResponseDto {
  @ApiProperty({
    description: 'List of jobs',
    type: [JobDto],
  })
  jobs: JobDto[]
}

@ApiSchema({ name: 'UpdateJobStatus' })
export class UpdateJobStatusDto {
  @ApiProperty({
    description: 'The new status of the job',
    enum: JobStatus,
    example: JobStatus.IN_PROGRESS,
  })
  @IsEnum(JobStatus)
  status: JobStatus

  @ApiPropertyOptional({
    description: 'Error message if the job failed',
    example: 'Failed to create sandbox',
  })
  @IsOptional()
  @IsString()
  errorMessage?: string
}
