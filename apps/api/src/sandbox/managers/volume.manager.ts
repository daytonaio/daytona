/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationBootstrap, OnApplicationShutdown, OnModuleInit } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, In } from 'typeorm'
import { Volume } from '../entities/volume.entity'
import { VolumeState } from '../enums/volume-state.enum'
import { Cron, CronExpression, SchedulerRegistry } from '@nestjs/schedule'
import { S3Client, CreateBucketCommand, ListBucketsCommand, PutBucketTaggingCommand } from '@aws-sdk/client-s3'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { TypedConfigService } from '../../config/typed-config.service'
import { deleteS3Bucket } from '../../common/utils/delete-s3-bucket'

import { TrackableJobExecutions } from '../../common/interfaces/trackable-job-executions'
import { TrackJobExecution } from '../../common/decorators/track-job-execution.decorator'
import { setTimeout } from 'timers/promises'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { ArchilClient, ArchilDiskMount } from '../services/archil/archil.client'
import { EncryptionService } from '../../encryption/encryption.service'
import { VOLUME_BACKEND_EXPERIMENTAL, VOLUME_BACKEND_S3FUSE } from '../services/volume.service'

const VOLUME_STATE_LOCK_KEY = 'volume-state-'

@Injectable()
export class VolumeManager
  implements OnModuleInit, TrackableJobExecutions, OnApplicationShutdown, OnApplicationBootstrap
{
  activeJobs = new Set<string>()

  private readonly logger = new Logger(VolumeManager.name)
  private processingVolumes: Set<string> = new Set()
  private skipTestConnection = false
  private s3Client: S3Client | null = null

  constructor(
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    private readonly configService: TypedConfigService,
    @InjectRedis() private readonly redis: Redis,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly schedulerRegistry: SchedulerRegistry,
    private readonly archilClient: ArchilClient,
    private readonly encryptionService: EncryptionService,
  ) {
    if (!this.configService.get('s3.endpoint')) {
      return
    }

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

  // Whether any volume backend is wired up. We start the cron processor as
  // long as at least one is configured; per-volume branching then routes to
  // the matching backend (and rejects volumes whose backend isn't set up).
  //
  // Note: the experimental backend always provisions an S3 bucket *and* an
  // Archil disk that mounts it, so it requires *both* services. Without S3
  // we still want the cron to run for s3fuse-only volumes, hence the OR.
  private get anyBackendConfigured(): boolean {
    return Boolean(this.s3Client) || this.archilClient.isConfigured()
  }

  async onModuleInit() {
    if (!this.s3Client) {
      return
    }

    if (this.skipTestConnection) {
      this.logger.debug('Skipping S3 connection test')
      return
    }

    await this.testConnection()
  }

  onApplicationBootstrap() {
    if (!this.anyBackendConfigured) {
      return
    }

    this.schedulerRegistry.getCronJob('process-pending-volumes').start()
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

  @Cron(CronExpression.EVERY_5_SECONDS, { name: 'process-pending-volumes', waitForCompletion: true, disabled: true })
  @TrackJobExecution()
  @LogExecution('process-pending-volumes')
  @WithInstrumentation()
  async processPendingVolumes() {
    if (!this.anyBackendConfigured) {
      return
    }

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

      const backend = volume.backend || VOLUME_BACKEND_S3FUSE
      if (backend === VOLUME_BACKEND_EXPERIMENTAL) {
        await this.provisionArchilDisk(volume, lockKey)
      } else {
        await this.provisionS3Bucket(volume, lockKey)
      }
      this.logger.debug(`Volume ${volume.id} created successfully (backend=${backend})`)
    } catch (error) {
      this.logger.error(`Error creating volume ${volume.id}:`, error)
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.ERROR,
        errorReason: error.message,
      })
    }
  }

  private async provisionS3Bucket(volume: Volume, lockKey: string): Promise<void> {
    if (!this.s3Client) {
      throw new Error(
        `Volume ${volume.id} uses backend 's3fuse' but S3 is not configured on this API. Configure S3 or change the volume's backend.`,
      )
    }

    await this.createAndTagS3Bucket(volume, lockKey)

    // Refresh lock before final state update
    await this.redis.setex(lockKey, 30, '1')

    await this.volumeRepository.save({
      ...volume,
      state: VolumeState.READY,
    })
  }

  // Creates the S3 bucket for `volume` and tags it. Idempotent enough for
  // our retry path: if the bucket already exists and we own it, AWS
  // returns BucketAlreadyOwnedByYou which we treat as success. Used by
  // both the s3fuse and experimental backends.
  private async createAndTagS3Bucket(volume: Volume, lockKey: string): Promise<void> {
    // Refresh lock before S3 operation
    await this.redis.setex(lockKey, 30, '1')

    try {
      await this.s3Client.send(new CreateBucketCommand({ Bucket: volume.getBucketName() }))
    } catch (error) {
      if (error?.name !== 'BucketAlreadyOwnedByYou') {
        throw error
      }
    }

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
  }

  // Builds the Archil mount configuration that backs an experimental
  // volume's disk with the Daytona-owned S3 bucket we just provisioned.
  //
  // We forward the *configured* S3 credentials to Archil. They grant access
  // to the entire S3 endpoint, but Archil only mounts the one named bucket
  // per disk so the practical blast radius is limited to that bucket. For
  // production deployments we recommend provisioning a least-privilege IAM
  // user scoped to `daytona-volume-*` and pointing `S3_ACCESS_KEY` /
  // `S3_SECRET_KEY` at it.
  private buildArchilMount(volume: Volume): ArchilDiskMount {
    const endpoint = this.configService.getOrThrow('s3.endpoint') as string
    const accessKeyId = this.configService.getOrThrow('s3.accessKey') as string
    const secretAccessKey = this.configService.getOrThrow('s3.secretKey') as string
    const endpointWithProtocol = endpoint.startsWith('http') ? endpoint : `https://${endpoint}`

    let isAws = false
    try {
      isAws = /(^|\.)amazonaws\.com$/i.test(new URL(endpointWithProtocol).hostname)
    } catch {
      // Malformed endpoint — fall through to s3-compatible, which carries
      // the raw URL and will surface the failure on Archil's side.
    }

    if (isAws) {
      return {
        type: 's3',
        bucketName: volume.getBucketName(),
        accessKeyId,
        secretAccessKey,
      }
    }
    return {
      type: 's3-compatible',
      bucketName: volume.getBucketName(),
      bucketEndpoint: endpointWithProtocol,
      accessKeyId,
      secretAccessKey,
    }
  }

  private async provisionArchilDisk(volume: Volume, lockKey: string): Promise<void> {
    this.assertArchilProvisioningPossible(volume)

    // Step 1: provision the S3 bucket that will back this disk. This is
    // the same operation as the s3fuse path so we share the helper.
    await this.createAndTagS3Bucket(volume, lockKey)

    // Step 2: ask Archil to create a disk that mounts the bucket. If this
    // fails we delete the bucket we just created so retries start from a
    // clean state (Archil's createDisk is idempotent on name, but we'd
    // rather not leave dangling buckets around if the user gives up).
    let withDisk: Volume
    try {
      await this.redis.setex(lockKey, 30, '1')
      withDisk = await this.ensureArchilDiskFor(volume)
    } catch (error) {
      this.logger.error(
        `Archil createDisk failed for volume ${volume.id}; rolling back the S3 bucket we just created`,
        error,
      )
      try {
        await deleteS3Bucket(this.s3Client, volume.getBucketName())
      } catch (rollbackError) {
        this.logger.error(
          `Rollback failed for volume ${volume.id}: bucket ${volume.getBucketName()} could not be deleted. Manual cleanup required.`,
          rollbackError,
        )
      }
      throw error
    }

    // Refresh lock before final state update
    await this.redis.setex(lockKey, 30, '1')
    await this.volumeRepository.save({ ...withDisk, state: VolumeState.READY })
  }

  // Idempotent provisioning of an Archil disk for the given volume. If the
  // volume row already has an `archilDiskId` and an encrypted token we
  // assume the disk exists and return as-is — useful both for retries on
  // the create path and for the migrate path where the row is already
  // READY but we're attaching an Archil disk on top of an existing bucket.
  //
  // Caveat: Archil's createDisk is idempotent on (name, configuration),
  // but mount tokens are one-time. If a previous attempt successfully
  // created the disk on Archil's side but we lost the response (e.g. DB
  // crash), the retry returns the same diskId without a new token and
  // surfaces an error. Recovery is manual: delete the disk by name on
  // Archil and retry. We never accept an empty token because that would
  // produce an unmountable disk.
  private async ensureArchilDiskFor(volume: Volume): Promise<Volume> {
    if (volume.archilDiskId && volume.archilMountTokenEnc) {
      return volume
    }
    this.assertArchilProvisioningPossible(volume)

    const created = await this.archilClient.createDisk({
      name: volume.getArchilDiskName(),
      region: this.archilClient.getDefaultRegion(),
      mount: this.buildArchilMount(volume),
    })

    const encryptedToken = await this.encryptionService.encrypt(created.mountToken)

    return await this.volumeRepository.save({
      ...volume,
      archilDiskId: created.diskId,
      archilRegion: created.region,
      archilMountTokenEnc: encryptedToken,
    })
  }

  private assertArchilProvisioningPossible(volume: Volume): void {
    if (!this.archilClient.isConfigured()) {
      throw new Error(
        `Volume ${volume.id} requires Archil but the control plane is not configured on this API. Set ARCHIL_API_KEY.`,
      )
    }
    if (!this.s3Client) {
      throw new Error(
        `Volume ${volume.id} requires Archil but S3 is not configured on this API. The experimental backend stores volume data in a Daytona-owned S3 bucket and requires S3 to be configured.`,
      )
    }
  }

  // Attaches an Archil disk to a volume that doesn't have one yet — used
  // by the migrate path to promote an existing s3fuse volume to the
  // experimental backend without re-creating its bucket. The bucket is
  // assumed to already exist from the volume's s3fuse era; we don't
  // recreate or re-tag it.
  //
  // Idempotent: calling this on a volume that already has an Archil disk
  // is a no-op.
  async attachArchilDiskTo(volume: Volume): Promise<Volume> {
    return await this.ensureArchilDiskFor(volume)
  }

  // Detaches and deletes the Archil disk associated with a volume — used
  // by the migrate path to demote an experimental volume back to s3fuse.
  // The S3 bucket is left intact so subsequent host-side mounts can read
  // the same data. Idempotent: calling this on a volume without an
  // Archil disk is a no-op.
  async detachArchilDiskFrom(volume: Volume): Promise<Volume> {
    if (volume.archilDiskId) {
      if (!this.archilClient.isConfigured()) {
        throw new Error(
          `Volume ${volume.id} has an Archil disk but the control plane is not configured on this API. Set ARCHIL_API_KEY to retry the migration.`,
        )
      }
      await this.archilClient.deleteDisk(
        volume.archilDiskId,
        volume.archilRegion || this.archilClient.getDefaultRegion(),
      )
    }
    return await this.volumeRepository.save({
      ...volume,
      archilDiskId: null,
      archilRegion: null,
      archilMountTokenEnc: null,
    })
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

      // Deletion is driven by the resources actually provisioned on the
      // row, not by `backend`. A volume that was migrated s3fuse →
      // experimental → s3fuse can have only a bucket; a volume currently
      // on `experimental` always has both. This way the migrate path
      // can't leak an Archil disk if the operator forgets to detach.
      if (volume.archilDiskId) {
        await this.deleteArchilDisk(volume, lockKey)
      }
      await this.deleteBackingBucket(volume, lockKey)

      // Refresh lock before final state update
      await this.redis.setex(lockKey, 30, '1')

      // Delete any existing volume record with the deleted state and the same name in the same organization
      await this.volumeRepository.delete({
        organizationId: volume.organizationId,
        name: `${volume.name}-deleted`,
        state: VolumeState.DELETED,
      })

      // Wipe the archil token from the row before marking it deleted so it
      // never lingers in plaintext-encrypted form on a deleted volume.
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.DELETED,
        name: `${volume.name}-deleted`,
        archilDiskId: null,
        archilRegion: null,
        archilMountTokenEnc: null,
      })
      this.logger.debug(`Volume ${volume.id} deleted successfully (backend=${volume.backend || VOLUME_BACKEND_S3FUSE})`)
    } catch (error) {
      this.logger.error(`Error deleting volume ${volume.id}:`, error)
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.ERROR,
        errorReason: error.message,
      })
    }
  }

  // Delete the volume's Archil disk. Always called *before*
  // deleteBackingBucket so the disk stops syncing into S3 — otherwise an
  // in-flight write could recreate objects right after we emptied the
  // bucket. Idempotent on missing disks (Archil returns 404 which is
  // treated as success in `archilClient.deleteDisk`).
  private async deleteArchilDisk(volume: Volume, lockKey: string): Promise<void> {
    if (!this.archilClient.isConfigured()) {
      throw new Error(
        `Volume ${volume.id} has an Archil disk but the control plane is not configured on this API. Set ARCHIL_API_KEY to retry deletion.`,
      )
    }

    await this.redis.setex(lockKey, 30, '1')
    await this.archilClient.deleteDisk(volume.archilDiskId, volume.archilRegion || this.archilClient.getDefaultRegion())
  }

  // Delete the volume's backing S3 bucket. Tolerates NoSuchBucket so
  // retries after a partial delete are safe.
  private async deleteBackingBucket(volume: Volume, lockKey: string): Promise<void> {
    if (!this.s3Client) {
      throw new Error(
        `Volume ${volume.id} has a backing S3 bucket but S3 is not configured on this API. Configure S3 to retry deletion.`,
      )
    }

    await this.redis.setex(lockKey, 30, '1')
    try {
      await deleteS3Bucket(this.s3Client, volume.getBucketName())
    } catch (error) {
      if (error?.name === 'NoSuchBucket') {
        this.logger.warn(`Bucket for volume ${volume.id} does not exist, treating as already deleted`)
      } else if (error?.name === 'BucketNotEmpty') {
        throw new Error('Volume deletion failed because the bucket is not empty. You may retry deletion.')
      } else {
        throw error
      }
    }
  }
}
