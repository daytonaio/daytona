/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnModuleInit } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, In } from 'typeorm'
import { Volume } from '../entities/volume.entity'
import { VolumeState } from '../enums/volume-state.enum'
import { Cron, CronExpression } from '@nestjs/schedule'
import { S3Client, CreateBucketCommand, ListBucketsCommand, PutBucketTaggingCommand } from '@aws-sdk/client-s3'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { TypedConfigService } from '../../config/typed-config.service'
import { deleteS3Bucket } from '../../common/utils/delete-s3-bucket'

const VOLUME_STATE_LOCK_KEY = 'volume-state-'

@Injectable()
export class VolumeManager implements OnModuleInit {
  private readonly logger = new Logger(VolumeManager.name)
  private processingVolumes: Set<string> = new Set()
  private skipTestConnection: boolean
  private s3Client: S3Client

  constructor(
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    private readonly configService: TypedConfigService,
    @InjectRedis() private readonly redis: Redis,
    private readonly redisLockProvider: RedisLockProvider,
  ) {
    const endpoint = this.configService.getOrThrow('s3.endpoint')
    const region = this.configService.getOrThrow('s3.region')
    const accessKeyId = this.configService.getOrThrow('s3.accessKey')
    const secretAccessKey = this.configService.getOrThrow('s3.secretKey')
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

  @Cron(CronExpression.EVERY_5_SECONDS)
  async processPendingVolumes() {
    try {
      // Lock the entire process
      const lockKey = 'process-pending-volumes'
      if (!(await this.redisLockProvider.lock(lockKey, 30))) {
        return
      }

      const pendingVolumes = await this.volumeRepository.find({
        where: {
          state: In([VolumeState.PENDING_CREATE, VolumeState.PENDING_DELETE]),
        },
      })

      await Promise.all(
        pendingVolumes.map(async (volume) => {
          if (this.processingVolumes.has(volume.id)) {
            return
          }

          // Get lock for this specific volume
          const volumeLockKey = `${VOLUME_STATE_LOCK_KEY}${volume.id}`
          const acquired = await this.redisLockProvider.lock(volumeLockKey, 30)
          if (!acquired) {
            return
          }

          try {
            this.processingVolumes.add(volume.id)
            await this.processVolumeState(volume)
          } finally {
            this.processingVolumes.delete(volume.id)
            await this.redisLockProvider.unlock(volumeLockKey)
          }
        }),
      )

      await this.redisLockProvider.unlock(lockKey)
    } catch (error) {
      this.logger.error('Error processing pending volumes:', error)
    }
  }

  private async processVolumeState(volume: Volume): Promise<void> {
    const volumeLockKey = `${VOLUME_STATE_LOCK_KEY}${volume.id}`

    try {
      switch (volume.state) {
        case VolumeState.PENDING_CREATE:
          await this.handlePendingCreate(volume, volumeLockKey)
          break
        case VolumeState.PENDING_DELETE:
          await this.handlePendingDelete(volume, volumeLockKey)
          break
      }
    } catch (error) {
      this.logger.error(`Error processing volume ${volume.id}:`, error)
      await this.volumeRepository.update(volume.id, {
        state: VolumeState.ERROR,
        errorReason: error.message,
      })
    }
  }

  private async handlePendingCreate(volume: Volume, lockKey: string): Promise<void> {
    try {
      // Refresh lock before state change
      await this.redis.setex(lockKey, 30, '1')

      // Update state to CREATING
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.CREATING,
      })

      // Refresh lock before S3 operation
      await this.redis.setex(lockKey, 30, '1')

      // Create bucket in Minio/S3
      const createBucketCommand = new CreateBucketCommand({
        Bucket: volume.getBucketName(),
      })

      await this.s3Client.send(createBucketCommand)

      await this.s3Client.send(
        new PutBucketTaggingCommand({
          Bucket: volume.getBucketName(),
          Tagging: {
            TagSet: [
              {
                Key: 'VolumeId',
                Value: volume.id,
              },
              {
                Key: 'OrganizationId',
                Value: volume.organizationId,
              },
              {
                Key: 'Environment',
                Value: this.configService.get('environment'),
              },
            ],
          },
        }),
      )

      // Refresh lock before final state update
      await this.redis.setex(lockKey, 30, '1')

      // Update volume state to READY
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.READY,
      })
      this.logger.debug(`Volume ${volume.id} created successfully`)
    } catch (error) {
      this.logger.error(`Error creating volume ${volume.id}:`, error)
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.ERROR,
        errorReason: error.message,
      })
    }
  }

  private async handlePendingDelete(volume: Volume, lockKey: string): Promise<void> {
    try {
      // Refresh lock before state change
      await this.redis.setex(lockKey, 30, '1')

      // Update state to DELETING
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.DELETING,
      })

      // Refresh lock before S3 operation
      await this.redis.setex(lockKey, 30, '1')

      // Delete bucket from Minio/S3
      await deleteS3Bucket(this.s3Client, volume.getBucketName())

      // Refresh lock before final state update
      await this.redis.setex(lockKey, 30, '1')

      // Delete any existing volume record with the deleted state and the same name in the same organization
      await this.volumeRepository.delete({
        organizationId: volume.organizationId,
        name: `${volume.name}-deleted`,
        state: VolumeState.DELETED,
      })

      // Update volume state to DELETED and rename
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.DELETED,
        name: `${volume.name}-deleted`,
      })
      this.logger.debug(`Volume ${volume.id} deleted successfully`)
    } catch (error) {
      this.logger.error(`Error deleting volume ${volume.id}:`, error)
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.ERROR,
        errorReason: error.message,
      })
    }
  }
}
