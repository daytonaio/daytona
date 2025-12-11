/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Post, Body, Param, Query, UseGuards, Logger, Req } from '@nestjs/common'
import { Request } from 'express'
import { ApiOAuth2, ApiTags, ApiOperation, ApiBearerAuth, ApiResponse, ApiParam, ApiQuery } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { RunnerAuthGuard } from '../../auth/runner-auth.guard'
import { RunnerContextDecorator } from '../../common/decorators/runner-context.decorator'
import { RunnerContext } from '../../common/interfaces/runner-context.interface'
import { JobDto, PollJobsResponseDto, UpdateJobStatusDto } from '../dto/job.dto'
import { JobService } from '../services/job.service'

@ApiTags('jobs')
@Controller('jobs')
@UseGuards(CombinedAuthGuard, RunnerAuthGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class JobController {
  private readonly logger = new Logger(JobController.name)

  constructor(private readonly jobService: JobService) {}

  @Get('poll')
  @ApiOperation({
    summary: 'Long poll for jobs',
    operationId: 'pollJobs',
    description:
      'Long poll endpoint for runners to fetch pending jobs. Returns immediately if jobs are available, otherwise waits up to timeout seconds.',
  })
  @ApiQuery({
    name: 'timeout',
    required: false,
    type: Number,
    description: 'Timeout in seconds for long polling (default: 30, max: 60)',
  })
  @ApiQuery({
    name: 'limit',
    required: false,
    type: Number,
    description: 'Maximum number of jobs to return (default: 10, max: 100)',
  })
  @ApiResponse({
    status: 200,
    description: 'List of jobs for the runner',
    type: PollJobsResponseDto,
  })
  async pollJobs(
    @Req() req: Request,
    @RunnerContextDecorator() runnerContext: RunnerContext,
    @Query('timeout') timeout?: number,
    @Query('limit') limit?: number,
  ): Promise<PollJobsResponseDto> {
    this.logger.debug(`Runner ${runnerContext.runnerId} polling for jobs (timeout: ${timeout}s, limit: ${limit})`)

    const timeoutSeconds = timeout ? Math.min(Number(timeout), 60) : 30
    const limitNumber = limit ? Math.min(Number(limit), 100) : 10

    // Create AbortSignal from request's 'close' event
    const abortController = new AbortController()
    const onClose = () => {
      this.logger.debug(`Runner ${runnerContext.runnerId} disconnected during polling, aborting`)
      abortController.abort()
    }
    req.on('close', onClose)

    try {
      const jobs = await this.jobService.pollJobs(
        runnerContext.runnerId,
        limitNumber,
        timeoutSeconds,
        abortController.signal,
      )
      this.logger.debug(`Returning ${jobs.length} jobs to runner ${runnerContext.runnerId}`)
      return { jobs }
    } catch (error) {
      if (abortController.signal.aborted) {
        this.logger.debug(`Polling aborted for disconnected runner ${runnerContext.runnerId}`)
        return { jobs: [] } // Return empty array on disconnect
      }
      this.logger.error(`Error polling jobs for runner ${runnerContext.runnerId}: ${error.message}`, error.stack)
      throw error
    } finally {
      req.off('close', onClose)
    }
  }

  @Get(':jobId')
  @ApiOperation({
    summary: 'Get job details',
    operationId: 'getJob',
  })
  @ApiParam({
    name: 'jobId',
    description: 'ID of the job',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Job details',
    type: JobDto,
  })
  async getJob(@RunnerContextDecorator() runnerContext: RunnerContext, @Param('jobId') jobId: string): Promise<JobDto> {
    this.logger.log(`Runner ${runnerContext.runnerId} fetching job ${jobId}`)

    const job = await this.jobService.findOne(jobId)
    if (!job) {
      throw new Error(`Job ${jobId} not found`)
    }

    return {
      id: job.id,
      type: job.type,
      status: job.status,
      resourceType: job.resourceType,
      resourceId: job.resourceId,
      payload: job.payload,
      errorMessage: job.errorMessage,
      createdAt: job.createdAt.toISOString(),
      updatedAt: job.updatedAt?.toISOString(),
    }
  }

  @Post(':jobId/status')
  @ApiOperation({
    summary: 'Update job status',
    operationId: 'updateJobStatus',
  })
  @ApiParam({
    name: 'jobId',
    description: 'ID of the job',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Job status updated successfully',
    type: JobDto,
  })
  async updateJobStatus(
    @RunnerContextDecorator() runnerContext: RunnerContext,
    @Param('jobId') jobId: string,
    @Body() updateJobStatusDto: UpdateJobStatusDto,
  ): Promise<JobDto> {
    this.logger.log(`Runner ${runnerContext.runnerId} updating job ${jobId} status to ${updateJobStatusDto.status}`)

    const job = await this.jobService.updateJobStatus(jobId, updateJobStatusDto.status, updateJobStatusDto.errorMessage)

    return {
      id: job.id,
      type: job.type,
      status: job.status,
      resourceType: job.resourceType,
      resourceId: job.resourceId,
      payload: job.payload,
      errorMessage: job.errorMessage,
      createdAt: job.createdAt.toISOString(),
      updatedAt: job.updatedAt?.toISOString(),
    }
  }
}
