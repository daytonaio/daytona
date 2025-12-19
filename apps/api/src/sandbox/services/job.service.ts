/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, LessThan, EntityManager } from 'typeorm'
import { Job } from '../entities/job.entity'
import { JobDto, JobStatus, JobType, ResourceType } from '../dto/job.dto'
import { ResourceTypeForJobType } from '../dto/job-type-map.dto'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { Cron, CronExpression } from '@nestjs/schedule'
import { JobStateHandlerService } from './job-state-handler.service'
import { propagation, context as otelContext } from '@opentelemetry/api'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'

@Injectable()
export class JobService {
  private readonly logger = new Logger(JobService.name)
  private readonly REDIS_JOB_QUEUE_PREFIX = 'runner:jobs:'

  constructor(
    @InjectRepository(Job)
    private readonly jobRepository: Repository<Job>,
    @InjectRedis() private readonly redis: Redis,
    private readonly jobStateHandlerService: JobStateHandlerService,
  ) {}

  /**
   * Create a job within the provided transaction manager
   * If manager is null, uses the default repository (for non-transactional operations)
   * @template T The JobType enum value - ensures compile-time type safety for resourceType and payload
   */
  async createJob<T extends JobType>(
    manager: EntityManager | null,
    type: T,
    runnerId: string,
    resourceType: ResourceTypeForJobType<T>,
    resourceId: string,
    payload?: string | Record<string, any>,
  ): Promise<Job> {
    // Use provided manager if available, otherwise use default repository
    const repo = manager ? manager.getRepository(Job) : this.jobRepository

    // Capture current OpenTelemetry trace context for distributed tracing
    const traceContext = this.captureTraceContext()

    const encodedPayload = typeof payload === 'string' ? payload : payload ? JSON.stringify(payload) : undefined

    const savedJob = await repo.save(
      new Job({
        type,
        runnerId,
        resourceType,
        resourceId,
        status: JobStatus.PENDING,
        payload: encodedPayload,
        traceContext,
      }),
    )

    // Log with context-specific info
    const contextInfo = resourceId ? `${resourceType} ${resourceId}` : 'N/A'

    this.logger.debug(`Created job ${savedJob.id} of type ${type} for ${contextInfo} on runner ${runnerId}`)

    // Notify runner via Redis - happens outside transaction
    // If transaction rolls back, notification is harmless (runner will poll and find nothing)
    await this.notifyRunner(runnerId, savedJob.id)

    return savedJob
  }

  private async notifyRunner(runnerId: string, jobId: string): Promise<void> {
    try {
      await this.redis.lpush(this.getRunnerQueueKey(runnerId), jobId)
      this.logger.debug(`Notified runner ${runnerId} about job ${jobId} via Redis`)
    } catch (error) {
      this.logger.warn(`Failed to notify runner ${runnerId} via Redis: ${error.message}`)
      // Job is still in DB, runner will pick it up via fallback polling
    }
  }

  async findOne(jobId: string): Promise<Job | null> {
    return this.jobRepository.findOneBy({ id: jobId })
  }

  async pollJobs(runnerId: string, limit = 10, timeoutSeconds = 30, abortSignal?: AbortSignal): Promise<JobDto[]> {
    const queueKey = this.getRunnerQueueKey(runnerId)
    const maxTimeout = Math.min(timeoutSeconds, 60) // Max 60 seconds

    // Check if already aborted
    if (abortSignal?.aborted) {
      this.logger.debug(`Poll request for runner ${runnerId} was already aborted`)
      return []
    }

    // STEP 1: Atomically claim pending jobs from database
    // This prevents duplicates by updating status to IN_PROGRESS
    let claimedJobs = await this.claimPendingJobs(runnerId, limit)

    if (claimedJobs.length > 0) {
      // Clear any stale job IDs from Redis queue
      try {
        await this.redis.del(queueKey)
      } catch (error) {
        this.logger.warn(`Failed to clear Redis queue: ${error.message}`)
      }

      return claimedJobs
    }

    // STEP 2: No existing jobs - wait for notification via Redis BRPOP
    // Create a new dedicated Redis client for this BRPOP to support concurrent polling from multiple runners
    // Each runner gets its own connection, preventing blocking issues
    let blockingClient: Redis | null = null
    try {
      this.logger.debug(`No existing jobs, runner ${runnerId} starting BRPOP with timeout ${maxTimeout}s`)

      blockingClient = this.redis.duplicate()

      // Wrap BRPOP in a promise that can be aborted
      const brpopPromise = blockingClient.brpop(queueKey, maxTimeout)

      let result: [string, string] | null = null
      if (abortSignal) {
        // Race between BRPOP and abort signal
        result = await Promise.race([
          brpopPromise,
          new Promise<null>((resolve) => {
            if (abortSignal.aborted) {
              resolve(null)
            } else {
              abortSignal.addEventListener('abort', () => resolve(null), { once: true })
            }
          }),
        ])

        // If aborted, disconnect immediately to cancel BRPOP
        if (abortSignal.aborted) {
          this.logger.debug(`BRPOP aborted for runner ${runnerId}, closing Redis connection`)
          blockingClient.disconnect()
          return []
        }
      } else {
        result = await brpopPromise
      }

      if (result) {
        // Got notification - job(s) available
        // Clear the entire queue (job IDs are just hints, not used directly)
        this.logger.debug(`Got notification from Redis for runner ${runnerId}`)

        try {
          await this.redis.del(queueKey)
        } catch (error) {
          this.logger.warn(`Failed to clear Redis queue: ${error.message}`)
        }

        // Atomically claim jobs from database
        claimedJobs = await this.claimPendingJobs(runnerId, limit)

        if (claimedJobs.length > 0) {
          this.logger.debug(`Claimed ${claimedJobs.length} jobs after Redis notification for runner ${runnerId}`)
          return claimedJobs
        }

        // Notification received but no jobs found - possible race condition
        this.logger.warn(`Received Redis notification but no pending jobs found for runner ${runnerId}`)
      } else {
        // BRPOP timeout - no jobs received
        this.logger.debug(`BRPOP timeout for runner ${runnerId}, no new jobs`)
      }
    } catch (error) {
      this.logger.error(`Redis BRPOP error for runner ${runnerId}: ${error.message}`)
      // Fall through to database polling fallback
    } finally {
      // Always close the blocking client to prevent connection leaks
      if (blockingClient) {
        try {
          await blockingClient.quit()
        } catch (error) {
          this.logger.warn(`Failed to close blocking Redis client: ${error.message}`)
        }
      }
    }

    // STEP 3: Final fallback - check database again
    // This handles race conditions and Redis failures
    claimedJobs = await this.claimPendingJobs(runnerId, limit)

    if (claimedJobs.length > 0) {
      this.logger.debug(`Claimed ${claimedJobs.length} pending jobs in fallback for runner ${runnerId}`)
    }

    return claimedJobs
  }

  async updateJobStatus(
    jobId: string,
    status: JobStatus,
    errorMessage?: string,
    resultMetadata?: string,
  ): Promise<Job> {
    const job = await this.findOne(jobId)
    if (!job) {
      throw new NotFoundException(`Job with ID ${jobId} not found`)
    }

    job.status = status
    if (errorMessage) {
      job.errorMessage = errorMessage
    }

    if (status === JobStatus.IN_PROGRESS && !job.startedAt) {
      job.startedAt = new Date()
    }

    if (status === JobStatus.COMPLETED || status === JobStatus.FAILED) {
      job.completedAt = new Date()
    }

    if (resultMetadata) {
      job.resultMetadata = resultMetadata
    }

    const updatedJob = await this.jobRepository.save(job)
    this.logger.debug(`Updated job ${jobId} status to ${status}`)

    // Handle job completion for v2 runners - update sandbox/snapshot/backup state
    if (status === JobStatus.COMPLETED || status === JobStatus.FAILED) {
      // Fire and forget - don't block the response
      this.jobStateHandlerService.handleJobCompletion(updatedJob).catch((error) => {
        this.logger.error(`Error handling job completion for job ${jobId}:`, error)
      })
    }

    return updatedJob
  }

  async findPendingJobsForRunner(runnerId: string, limit = 10): Promise<Job[]> {
    return this.jobRepository.find({
      where: {
        runnerId,
        status: JobStatus.PENDING,
      },
      order: {
        createdAt: 'ASC',
      },
      take: limit,
    })
  }

  async findJobsForRunner(runnerId: string, status?: JobStatus, page = 1, limit = 100): Promise<PaginatedList<JobDto>> {
    const whereCondition: { runnerId: string; status?: JobStatus } = { runnerId }

    if (status) {
      whereCondition.status = status
    }

    const [jobs, total] = await this.jobRepository.findAndCount({
      where: whereCondition,
      order: {
        createdAt: 'DESC',
      },
      skip: (page - 1) * limit,
      take: limit,
    })

    return {
      items: jobs.map((job) => new JobDto(job)),
      total,
      page,
      totalPages: Math.ceil(total / limit),
    }
  }

  async findJobsBySandboxId(sandboxId: string): Promise<Job[]> {
    return this.findJobsByResourceId(ResourceType.SANDBOX, sandboxId)
  }

  async findJobsByResourceId(resourceType: ResourceType, resourceId: string): Promise<Job[]> {
    return this.jobRepository.find({
      where: {
        resourceType,
        resourceId,
      },
      order: {
        createdAt: 'DESC',
      },
    })
  }

  /**
   * Captures the current OpenTelemetry trace context in W3C Trace Context format
   * This allows distributed tracing across the API and runner services
   * @returns A map of trace context headers (traceparent, tracestate)
   */
  private captureTraceContext(): Record<string, string> | null {
    try {
      const carrier: Record<string, string> = {}

      // Extract current trace context into carrier object using W3C Trace Context format
      propagation.inject(otelContext.active(), carrier)

      // Return the carrier if it contains trace information
      if (Object.keys(carrier).length > 0) {
        this.logger.debug(`Captured trace context: ${JSON.stringify(carrier)}`)
        return carrier
      }
    } catch (error) {
      this.logger.warn(`Failed to capture trace context: ${error.message}`)
    }

    return null
  }

  private getRunnerQueueKey(runnerId: string): string {
    return `${this.REDIS_JOB_QUEUE_PREFIX}${runnerId}`
  }

  /**
   * Cron job to check for stale jobs and mark them as failed
   * Runs every minute to find jobs that have been IN_PROGRESS for too long
   */
  @Cron(CronExpression.EVERY_MINUTE)
  async handleStaleJobs(): Promise<void> {
    const staleThresholdMinutes = 10
    const staleThreshold = new Date(Date.now() - staleThresholdMinutes * 60 * 1000)

    try {
      // Find jobs that are IN_PROGRESS but haven't been updated in the threshold time
      const staleJobs = await this.jobRepository.find({
        where: {
          status: JobStatus.IN_PROGRESS,
          updatedAt: LessThan(staleThreshold),
        },
      })

      if (staleJobs.length === 0) {
        return
      }

      this.logger.warn(`Found ${staleJobs.length} stale jobs, marking as failed`)

      // Mark each stale job as failed with timeout error
      for (const job of staleJobs) {
        try {
          await this.updateJobStatus(
            job.id,
            JobStatus.FAILED,
            `Job timed out - no update received for ${staleThresholdMinutes} minutes`,
          )

          this.logger.warn(
            `Marked job ${job.id} (type: ${job.type}, resource: ${job.resourceType} ${job.resourceId}) as failed due to timeout`,
          )
        } catch (error) {
          this.logger.error(`Error marking job ${job.id} as failed: ${error.message}`, error.stack)
        }
      }
    } catch (error) {
      this.logger.error(`Error handling stale jobs: ${error.message}`, error.stack)
    }
  }

  /**
   * Atomically claim pending jobs by updating their status to IN_PROGRESS
   * This prevents duplicate processing of the same job
   */
  private async claimPendingJobs(runnerId: string, limit: number): Promise<JobDto[]> {
    // Find pending jobs
    const jobs = await this.jobRepository.find({
      where: {
        runnerId,
        status: JobStatus.PENDING,
      },
      order: {
        createdAt: 'ASC',
      },
      take: limit,
    })

    if (jobs.length === 0) {
      return []
    }

    // Update jobs to IN_PROGRESS
    const now = new Date()
    const claimedJobs: JobDto[] = []

    for (const job of jobs) {
      try {
        job.status = JobStatus.IN_PROGRESS
        job.startedAt = now
        job.updatedAt = now

        // save() with @VersionColumn will automatically check version and throw OptimisticLockVersionMismatchError if changed
        const savedJob = await this.jobRepository.save(job)

        claimedJobs.push(new JobDto(savedJob))
      } catch (error) {
        // If optimistic lock fails, job was already claimed by another runner - skip it
        this.logger.debug(`Job ${job.id} already claimed by another runner (version mismatch)`)
      }
    }

    if (claimedJobs.length > 0) {
      this.logger.debug(`Claimed ${claimedJobs.length} existing pending jobs for runner ${runnerId}`)
    }

    return claimedJobs
  }
}
