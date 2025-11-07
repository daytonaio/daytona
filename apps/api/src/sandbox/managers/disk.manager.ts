/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationShutdown, OnModuleInit } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, In } from 'typeorm'
import { Disk } from '../entities/disk.entity'
import { DiskState } from '../enums/disk-state.enum'
import { Cron, CronExpression } from '@nestjs/schedule'
import { LessThan } from 'typeorm'
import { S3Client, ListBucketsCommand, CopyObjectCommand } from '@aws-sdk/client-s3'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { TypedConfigService } from '../../config/typed-config.service'
import { deleteS3Folder } from '../../common/utils/delete-s3-folder'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { RunnerService } from '../services/runner.service'

import { TrackableJobExecutions } from '../../common/interfaces/trackable-job-executions'
import { TrackJobExecution } from '../../common/decorators/track-job-execution.decorator'
import { setTimeout } from 'timers/promises'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { OnEvent } from '@nestjs/event-emitter'
import { DiskEvents } from '../constants/disk-events'
import { DiskCreatedEvent } from '../events/disk-created.event'

const DISK_STATE_LOCK_KEY = 'disk-state-'

@Injectable()
export class DiskManager implements OnModuleInit, TrackableJobExecutions, OnApplicationShutdown {
  activeJobs = new Set<string>()

  private readonly logger = new Logger(DiskManager.name)
  private processingDisks: Set<string> = new Set()
  private skipTestConnection: boolean
  private s3Client: S3Client
  private s3Bucket: string

  constructor(
    @InjectRepository(Disk)
    private readonly diskRepository: Repository<Disk>,
    private readonly configService: TypedConfigService,
    @InjectRedis() private readonly redis: Redis,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    private readonly runnerService: RunnerService,
  ) {
    const endpoint = this.configService.getOrThrow('s3.endpoint')
    const region = this.configService.getOrThrow('s3.region')
    const accessKeyId = this.configService.getOrThrow('s3.accessKey')
    const secretAccessKey = this.configService.getOrThrow('s3.secretKey')
    this.s3Bucket = this.configService.getOrThrow('s3.defaultBucket')
    this.skipTestConnection = this.configService.get('skipConnections')

    this.s3Client = new S3Client({
      endpoint: endpoint.startsWith('http') ? endpoint : `http://${endpoint}`,
      region,
      credentials: {
        accessKeyId,
        secretAccessKey,
      },
      forcePathStyle: true,
    })
  }

  async onModuleInit() {
    if (this.skipTestConnection) {
      this.logger.debug('Skipping S3 connection test')
      return
    }

    await this.testConnection()
  }

  async onApplicationShutdown() {
    //  wait for all active jobs to finish
    while (this.activeJobs.size > 0) {
      this.logger.log(`Waiting for ${this.activeJobs.size} active jobs to finish`)
      await setTimeout(1000)
    }
  }

  private async testConnection() {
    try {
      // Try a simple operation to test the connection
      const command = new ListBucketsCommand({})
      await this.s3Client.send(command)
      this.logger.debug('Successfully connected to S3')
    } catch (error) {
      this.logger.error('Failed to connect to S3:', error)
      throw error
    }
  }

  @Cron(CronExpression.EVERY_5_SECONDS, { name: 'process-pending-disks', waitForCompletion: true })
  @TrackJobExecution()
  @LogExecution('process-pending-disks')
  async processPendingDisks() {
    try {
      // Lock the entire process
      const lockKey = 'process-pending-disks'
      if (!(await this.redisLockProvider.lock(lockKey, 30))) {
        return
      }

      const pendingDisks = await this.diskRepository.find({
        where: {
          state: In([DiskState.PENDING_DELETE, DiskState.PENDING_PUSH, DiskState.PUSHING, DiskState.FORKING]),
        },
      })

      await Promise.all(
        pendingDisks.map(async (disk) => {
          if (this.processingDisks.has(disk.id)) {
            return
          }

          try {
            this.processingDisks.add(disk.id)
            await this.processDiskState(disk)
          } finally {
            this.processingDisks.delete(disk.id)
          }
        }),
      )

      await this.redisLockProvider.unlock(lockKey)
    } catch (error) {
      this.logger.error('Error processing pending disks:', error)
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'push-detached-disks', waitForCompletion: true })
  @TrackJobExecution()
  @LogExecution('push-detached-disks')
  async pushDetachedDisks() {
    try {
      // Lock the entire process
      const lockKey = 'push-detached-disks'
      if (!(await this.redisLockProvider.lock(lockKey, 30))) {
        return
      }

      // Calculate timestamp for 10 seconds ago
      const tenSecondsAgo = new Date(Date.now() - 10 * 1000)

      // Query disks with state DETACHED and updatedAt older than 10 seconds
      const detachedDisks = await this.diskRepository.find({
        where: {
          state: DiskState.DETACHED,
          updatedAt: LessThan(tenSecondsAgo),
        },
      })

      this.logger.debug(`Found ${detachedDisks.length} detached disks to push`)

      await Promise.all(
        detachedDisks.map(async (disk) => {
          if (this.processingDisks.has(disk.id)) {
            return
          }

          // Get lock for this specific disk
          const diskLockKey = `${DISK_STATE_LOCK_KEY}${disk.id}`
          const acquired = await this.redisLockProvider.lock(diskLockKey, 30)
          if (!acquired) {
            return
          }

          this.processingDisks.add(disk.id)
          console.log(`### Pushing detached disk ${disk.id} ###`)
          await this.pushDetachedDisk(disk)
        }),
      )

      await this.redisLockProvider.unlock(lockKey)
    } catch (error) {
      this.logger.error('Error pushing detached disks:', error)
    }
  }

  private async processDiskState(disk: Disk): Promise<void> {
    // Get lock for this specific disk
    const diskLockKey = `${DISK_STATE_LOCK_KEY}${disk.id}`
    const acquired = await this.redisLockProvider.lock(diskLockKey, 30)
    if (!acquired) {
      return
    }

    try {
      switch (disk.state) {
        case DiskState.PUSHING:
          await this.redis.setex(diskLockKey, 30, '1')
          await this.handlePushing(disk)
          break
        case DiskState.PENDING_PUSH:
          await this.redis.setex(diskLockKey, 30, '1')
          await this.handlePendingPush(disk)
          break
        case DiskState.PENDING_DELETE:
          await this.handlePendingDelete(disk)
          break
        case DiskState.FORKING:
          await this.handleForking(disk)
          break
      }
    } catch (error) {
      this.logger.error(`Error processing disk ${disk.id}:`, error)
      await this.diskRepository.save({
        ...disk,
        state: DiskState.ERROR,
        errorReason: error.message,
      })
    } finally {
      await this.redisLockProvider.unlock(diskLockKey)
    }
  }

  // when forking a stored disk, we just need to duplicate the disk in S3
  private async handleForkingStored(disk: Disk): Promise<void> {
    const baseDiskPrefix = `disks/${disk.baseDiskId}`
    const diskFolderPrefix = `disks/${disk.id}`
    // clone the disk folder from the base disk to the new disk
    await this.s3Client.send(
      new CopyObjectCommand({
        Bucket: this.s3Bucket,
        CopySource: `${this.s3Bucket}/${baseDiskPrefix}`,
        Key: diskFolderPrefix,
      }),
    )

    // update the disk state to STORED
    // as the fork operation is complete when the base disk is stored
    disk.state = DiskState.STORED
    await this.diskRepository.save(disk)
  }

  // when forking a detached disk, we need to push the disk to the runner
  private async handleForkingLocked(disk: Disk): Promise<void> {
    try {
      const runner = await this.runnerService.findOne(disk.runnerId)
      if (!runner) {
        throw new Error(`Runner ${disk.runnerId} not found for disk ${disk.id}`)
      }

      // Create runner adapter and initiate disk push
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)
      const diskInfo = await runnerAdapter.getDiskInfo(disk.baseDiskId)
      if (!diskInfo) {
        throw new Error(`Disk info not found for base disk ${disk.baseDiskId}`)
      }

      await runnerAdapter.forkDisk(disk.baseDiskId, disk.id)

      disk.state = DiskState.DETACHED
      await this.diskRepository.save(disk)

      const existingDisk = await this.diskRepository.findOne({
        where: { id: disk.baseDiskId },
      })
      if (!existingDisk) {
        throw new Error(`Forked disk ${disk.id} not found after fork operation`)
      }
      existingDisk.state = DiskState.DETACHED
      await this.diskRepository.save(existingDisk)

      this.logger.debug(`Forked disk ${existingDisk.id} from ${existingDisk.baseDiskId} on runner ${runner.domain}`)
    } catch (error) {
      this.logger.error(`Error uploading disk ${disk.id}:`, error)
      await this.diskRepository.save({
        ...disk,
        state: DiskState.ERROR,
        errorReason: error.message,
      })
    }
  }

  private async handleForking(disk: Disk): Promise<void> {
    const baseDisk = await this.diskRepository.findOne({
      where: { id: disk.baseDiskId },
    })
    if (!baseDisk) {
      throw new Error(`Base disk ${disk.baseDiskId} not found for fork ${disk.id}`)
    }

    switch (baseDisk.state) {
      case DiskState.STORED:
        await this.handleForkingStored(disk)
        break
      case DiskState.LOCKED:
        // locked disks are already on the runner
        // these are disks in detached state that are being forked
        await this.handleForkingLocked(disk)
        break
      default:
        throw new Error(`Base disk ${disk.baseDiskId} is in invalid state ${baseDisk.state} for fork ${disk.id}`)
    }
  }

  private async handlePushing(disk: Disk): Promise<void> {
    try {
      const runner = await this.runnerService.findOne(disk.runnerId)
      if (!runner) {
        throw new Error(`Runner ${disk.runnerId} not found for disk ${disk.id}`)
      }

      // Create runner adapter and initiate disk push
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)
      const diskInfo = await runnerAdapter.getDiskInfo(disk.id)
      if (!diskInfo) {
        throw new Error(`Disk info not found for disk ${disk.id}`)
      }

      if (diskInfo.inS3) {
        // Update disk state to PUSHING
        await this.diskRepository.save({
          ...disk,
          state: DiskState.STORED,
        })
      }
    } catch (error) {
      this.logger.error(`Error uploading disk ${disk.id}:`, error)
      await this.diskRepository.save({
        ...disk,
        state: DiskState.ERROR,
        errorReason: error.message,
      })
    }
  }

  private async handlePendingDelete(disk: Disk): Promise<void> {
    try {
      // Update state to DELETING
      await this.diskRepository.save({
        ...disk,
        state: DiskState.DELETING,
      })

      // Delete disk folder from S3
      // Note, this will leave the layers in S3
      // The layer deletion is handled by the external layer manager service
      const diskFolderPrefix = `disks/${disk.id}`
      await deleteS3Folder(this.s3Client, this.s3Bucket, diskFolderPrefix)

      if (DiskState.DETACHED) {
        // If the disk is still on the runner, we need to tell the runner to delete the disk
        const runner = await this.runnerService.findOne(disk.runnerId)
        if (!runner) {
          throw new Error(`Runner ${disk.runnerId} not found for disk ${disk.id}`)
        }
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)
        await runnerAdapter.deleteDisk(disk.id)
      }

      // Delete any existing disk record with the deleted state and the same name in the same organization
      await this.diskRepository.delete({
        organizationId: disk.organizationId,
        name: `${disk.name}-deleted`,
        state: DiskState.DELETED,
      })

      // Update disk state to DELETED and rename
      await this.diskRepository.save({
        ...disk,
        state: DiskState.DELETED,
        name: `${disk.name}-deleted`,
      })
      this.logger.debug(`Disk ${disk.id} deleted successfully`)
    } catch (error) {
      this.logger.error(`Error deleting disk ${disk.id}:`, error)
      await this.diskRepository.save({
        ...disk,
        state: DiskState.ERROR,
        errorReason: error.message,
      })
    }
  }

  private async handlePendingPush(disk: Disk): Promise<void> {
    try {
      // Get the runner for this disk
      if (!disk.runnerId) {
        throw new Error(`Disk ${disk.id} has no runner ID`)
      }

      const runner = await this.runnerService.findOne(disk.runnerId)
      if (!runner) {
        throw new Error(`Runner ${disk.runnerId} not found for disk ${disk.id}`)
      }

      // Create runner adapter and initiate disk upload
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)
      await runnerAdapter.pushDisk(disk.id)

      // Update disk state to STORED (successfully uploaded)
      await this.diskRepository.save({
        ...disk,
        state: DiskState.PUSHING,
      })

      this.logger.debug(`Disk ${disk.id} push started`)
    } catch (error) {
      this.logger.error(`Error starting push for disk ${disk.id}:`, error)
      await this.diskRepository.save({
        ...disk,
        state: DiskState.ERROR,
        errorReason: error.message,
      })
    }
  }

  private async pushDetachedDisk(disk: Disk): Promise<void> {
    let runner = null
    try {
      // Get the runner for this disk
      if (!disk.runnerId) {
        throw new Error(`Disk ${disk.id} has no runner ID`)
      }

      runner = await this.runnerService.findOne(disk.runnerId)
      if (!runner) {
        throw new Error(`Runner ${disk.runnerId} not found for disk ${disk.id}`)
      }

      this.logger.debug(`Starting push for detached disk ${disk.id} on runner ${runner.domain}`)

      // Create runner adapter and check disk info first
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      // Get disk info to validate it exists and is in correct state
      let diskInfo
      try {
        diskInfo = await runnerAdapter.getDiskInfo(disk.id)
        this.logger.debug(`Disk info for ${disk.id}:`, {
          name: diskInfo.name,
          sizeGB: diskInfo.sizeGB,
          isMounted: diskInfo.isMounted,
          inS3: diskInfo.inS3,
        })
      } catch (infoError) {
        this.logger.warn(`Could not get disk info for ${disk.id}:`, infoError.message)
        // Continue with push attempt even if info retrieval fails
      }

      // Check if disk is already in S3
      if (diskInfo?.inS3) {
        this.logger.debug(`Disk ${disk.id} is already in S3, updating state to STORED`)
        await this.diskRepository.save({
          ...disk,
          state: DiskState.STORED,
        })
        return
      }

      // Attempt to push the disk
      await runnerAdapter.pushDisk(disk.id)

      // Update disk state to PENDING_PUSH (push operation initiated)
      await this.diskRepository.save({
        ...disk,
        state: DiskState.PENDING_PUSH,
      })

      this.logger.debug(`Detached disk ${disk.id} push started successfully`)
    } catch (error) {
      this.logger.error(`Error starting push for detached disk ${disk.id}:`, {
        error: error.message,
        stack: error.stack,
        diskId: disk.id,
        runnerId: disk.runnerId,
        runnerDomain: runner?.domain,
      })
      await this.diskRepository.save({
        ...disk,
        state: DiskState.ERROR,
        errorReason: error.message,
      })
    }
  }

  @OnEvent(DiskEvents.CREATED)
  @TrackJobExecution()
  private async handleSandboxCreatedEvent(event: DiskCreatedEvent) {
    this.processDiskState(event.disk).catch(this.logger.error)
  }
}
