/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional, ApiSchema } from '@nestjs/swagger'
import { IsEnum, IsObject, IsOptional, IsString } from 'class-validator'
import { JobType } from '../enums/job-type.enum'
import { JobStatus } from '../enums/job-status.enum'
import { ResourceType } from '../enums/resource-type.enum'
import { Job } from '../entities/job.entity'
import { PageNumber } from '../../common/decorators/page-number.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'

// Re-export enums for convenience
export { JobType, JobStatus, ResourceType }

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
    enumName: 'JobType',
    example: JobType.CREATE_SANDBOX,
  })
  @IsEnum(JobType)
  type: JobType

  @ApiProperty({
    description: 'The status of the job',
    enum: JobStatus,
    enumName: 'JobStatus',
    example: JobStatus.PENDING,
  })
  @IsEnum(JobStatus)
  status: JobStatus

  @ApiProperty({
    description: 'The type of resource this job operates on',
    enum: ResourceType,
    example: ResourceType.SANDBOX,
  })
  @IsEnum(ResourceType)
  resourceType: ResourceType

  @ApiProperty({
    description: 'The ID of the resource this job operates on (sandboxId, snapshotRef, etc.)',
    example: 'sandbox123',
  })
  @IsString()
  resourceId: string

  @ApiPropertyOptional({
    description: 'Job-specific JSON-encoded payload data (operational metadata)',
  })
  @IsOptional()
  payload?: string

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

  constructor(job: Job) {
    this.id = job.id
    this.type = job.type
    this.status = job.status
    this.resourceType = job.resourceType
    this.resourceId = job.resourceId
    this.payload = job.payload || undefined
    this.traceContext = job.traceContext || undefined
    this.errorMessage = job.errorMessage || undefined
    this.createdAt = job.createdAt.toISOString()
    this.updatedAt = job.updatedAt?.toISOString()
  }
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

@ApiSchema({ name: 'PaginatedJobs' })
export class PaginatedJobsDto {
  @ApiProperty({ type: [JobDto] })
  items: JobDto[]

  @ApiProperty()
  total: number

  @ApiProperty()
  page: number

  @ApiProperty()
  totalPages: number
}

@ApiSchema({ name: 'ListJobsQuery' })
export class ListJobsQueryDto {
  @PageNumber(1)
  page = 1

  @PageLimit(100)
  limit = 100

  @ApiPropertyOptional({
    description: 'Filter by job status',
    enum: JobStatus,
    enumName: 'JobStatus',
    example: JobStatus.PENDING,
  })
  @IsOptional()
  @IsEnum(JobStatus)
  status?: JobStatus
}

@ApiSchema({ name: 'UpdateJobStatus' })
export class UpdateJobStatusDto {
  @ApiProperty({
    description: 'The new status of the job',
    enum: JobStatus,
    enumName: 'JobStatus',
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

  @ApiPropertyOptional({
    description: 'Result metadata for the job',
  })
  @IsOptional()
  @IsString()
  resultMetadata?: string
}
