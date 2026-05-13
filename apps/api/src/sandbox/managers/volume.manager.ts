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
import {
  S3Client,
  CreateBucketCommand,
  DeleteObjectsCommand,
  ListBucketsCommand,
  ListObjectsV2Command,
  PutBucketTaggingCommand,
} from '@aws-sdk/client-s3'
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
import { LayeredVolumeClient, LayeredDiskMount } from '../services/layered/layered-volume.client'
import { VOLUME_BACKEND_LAYERED, VOLUME_BACKEND_S3FUSE } from '../services/volume.service'
import { Organization } from '../../organization/entities/organization.entity'

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
    @InjectRepository(Organization)
    private readonly organizationRepository: Repository<Organization>,
    private readonly configService: TypedConfigService,
    @InjectRedis() private readonly redis: Redis,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly schedulerRegistry: SchedulerRegistry,
    private readonly layeredClient: LayeredVolumeClient,
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
  // Note: the layered backend always provisions an S3 bucket *and* a
  // layered disk that mounts it, so it requires *both* services. Without S3
  // we still want the cron to run for s3fuse-only volumes, hence the OR.
  private get anyBackendConfigured(): boolean {
    return Boolean(this.s3Client) || this.layeredClient.isConfigured()
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
      if (backend === VOLUME_BACKEND_LAYERED) {
        await this.provisionLayeredDisk(volume, lockKey)
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

    await this.createAndTagPerVolumeS3Bucket(volume, lockKey)

    // Refresh lock before final state update
    await this.redis.setex(lockKey, 30, '1')

    await this.volumeRepository.save({
      ...volume,
      state: VolumeState.READY,
    })
  }

  // Creates the per-volume s3fuse bucket for `volume` and tags it.
  // Idempotent enough for our retry path: if the bucket already exists and
  // we own it, AWS returns BucketAlreadyOwnedByYou which we treat as
  // success.
  private async createAndTagPerVolumeS3Bucket(volume: Volume, lockKey: string): Promise<void> {
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
            { Key: 'VolumeId', Value: volume.id },
            { Key: 'OrganizationId', Value: volume.organizationId },
            { Key: 'Environment', Value: this.configService.get('environment') },
          ],
        },
      }),
    )
  }

  // Idempotent provisioning of the per-organization layered bucket. The
  // first layered volume in an organization lazily creates the bucket and
  // persists the chosen name on `organization.layeredBucketName`; later
  // volumes reuse it.
  private async ensureOrgLayeredBucket(volume: Volume): Promise<string> {
    if (!this.s3Client) {
      throw new Error(
        `Volume ${volume.id} requires layered storage but S3 is not configured on this API. The layered backend stores volume data in a Daytona-owned S3 bucket and requires S3 to be configured.`,
      )
    }
    if (!volume.organizationId) {
      throw new Error(`Volume ${volume.id} has no organizationId; cannot provision a per-organization layered bucket.`)
    }

    const org = await this.organizationRepository.findOne({ where: { id: volume.organizationId } })
    if (!org) {
      throw new Error(`Organization ${volume.organizationId} not found while provisioning layered volume ${volume.id}`)
    }

    const bucketName = org.layeredBucketName || `daytona-org-volumes-${org.id}`

    try {
      await this.s3Client.send(new CreateBucketCommand({ Bucket: bucketName }))
    } catch (error) {
      if (error?.name !== 'BucketAlreadyOwnedByYou') {
        throw error
      }
    }

    await this.s3Client.send(
      new PutBucketTaggingCommand({
        Bucket: bucketName,
        Tagging: {
          TagSet: [
            { Key: 'OrganizationId', Value: org.id },
            { Key: 'Purpose', Value: 'layered-volumes' },
            { Key: 'Environment', Value: this.configService.get('environment') },
          ],
        },
      }),
    )

    if (org.layeredBucketName !== bucketName) {
      org.layeredBucketName = bucketName
      await this.organizationRepository.save(org)
    }

    return bucketName
  }

  // Builds the layered mount configuration that backs a volume's disk with
  // the per-organization S3 bucket + the volume's prefix folder inside it.
  //
  // We forward the *configured* S3 credentials to the control plane. They
  // grant access to the entire S3 endpoint, but the disk is created with
  // `bucketPrefix = <volumeId>/` so the practical blast radius is limited
  // to that folder. For production deployments we recommend provisioning a
  // least-privilege IAM user scoped to `daytona-org-volumes-*` and pointing
  // `S3_ACCESS_KEY` / `S3_SECRET_KEY` at it.
  private buildLayeredMount(bucketName: string, volume: Volume): LayeredDiskMount {
    const endpoint = this.configService.getOrThrow('s3.endpoint') as string
    const accessKeyId = this.configService.getOrThrow('s3.accessKey') as string
    const secretAccessKey = this.configService.getOrThrow('s3.secretKey') as string
    const endpointWithProtocol = endpoint.startsWith('http') ? endpoint : `https://${endpoint}`

    let isAws = false
    try {
      isAws = /(^|\.)amazonaws\.com$/i.test(new URL(endpointWithProtocol).hostname)
    } catch {
      // Malformed endpoint — fall through to s3-compatible, which carries
      // the raw URL and will surface the failure on the control plane side.
    }

    if (isAws) {
      return {
        type: 's3',
        bucketName,
        bucketPrefix: volume.getLayeredBucketPrefix(),
        accessKeyId,
        secretAccessKey,
      }
    }
    return {
      type: 's3-compatible',
      bucketName,
      bucketEndpoint: endpointWithProtocol,
      bucketPrefix: volume.getLayeredBucketPrefix(),
      accessKeyId,
      secretAccessKey,
    }
  }

  private async provisionLayeredDisk(volume: Volume, lockKey: string): Promise<void> {
    this.assertLayeredProvisioningPossible(volume)

    // Step 1: provision (or reuse) the per-org bucket that backs every
    // layered volume in this organization.
    await this.redis.setex(lockKey, 30, '1')
    const bucketName = await this.ensureOrgLayeredBucket(volume)

    // Step 2: ask the control plane to create a disk that mounts the
    // bucket scoped to this volume's prefix. The disk has its own
    // generated token at creation; we throw it away — per-sandbox tokens
    // are minted lazily on the `sandbox_volume` table. The org bucket is
    // intentionally not rolled back on failure: it's a shared resource
    // and the next volume create will reuse it.
    await this.redis.setex(lockKey, 30, '1')
    let updated: Volume
    try {
      updated = await this.ensureLayeredDiskFor(volume, bucketName)
    } catch (error) {
      this.logger.error(`Layered createDisk failed for volume ${volume.id}`, error)
      throw error
    }

    // Refresh lock before final state update
    await this.redis.setex(lockKey, 30, '1')
    await this.volumeRepository.save({ ...updated, state: VolumeState.READY })
  }

  // Idempotent provisioning of a layered disk for the given volume. If the
  // volume row already has a `layeredDiskId` we assume the disk exists and
  // return as-is — useful for retries after a partial provision.
  //
  // Caveat: the control plane's createDisk is idempotent on
  // (name, configuration), but its initial token is one-time. We
  // intentionally discard that initial token; per-sandbox tokens are
  // minted on first attach via `LayeredVolumeClient.mintMountKey`.
  private async ensureLayeredDiskFor(volume: Volume, bucketName: string): Promise<Volume> {
    if (volume.layeredDiskId) {
      return volume
    }

    const created = await this.layeredClient.createDisk({
      name: volume.getLayeredDiskName(),
      region: this.layeredClient.getDefaultRegion(),
      mount: this.buildLayeredMount(bucketName, volume),
    })

    return await this.volumeRepository.save({
      ...volume,
      layeredDiskId: created.diskId,
      layeredRegion: created.region,
    })
  }

  private assertLayeredProvisioningPossible(volume: Volume): void {
    if (!this.layeredClient.isConfigured()) {
      throw new Error(
        `Volume ${volume.id} requires the layered backend but its control plane is not configured on this API. Set LAYERED_API_KEY.`,
      )
    }
    if (!this.s3Client) {
      throw new Error(
        `Volume ${volume.id} requires the layered backend but S3 is not configured on this API. The layered backend stores volume data in a Daytona-owned S3 bucket and requires S3 to be configured.`,
      )
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

      const backend = volume.backend || VOLUME_BACKEND_S3FUSE
      if (backend === VOLUME_BACKEND_LAYERED) {
        await this.deleteLayeredDisk(volume, lockKey)
        await this.deleteLayeredBucketPrefix(volume, lockKey)
      } else {
        await this.deletePerVolumeBucket(volume, lockKey)
      }

      // Refresh lock before final state update
      await this.redis.setex(lockKey, 30, '1')

      // Delete any existing volume record with the deleted state and the same name in the same organization
      await this.volumeRepository.delete({
        organizationId: volume.organizationId,
        name: `${volume.name}-deleted`,
        state: VolumeState.DELETED,
      })

      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.DELETED,
        name: `${volume.name}-deleted`,
        layeredDiskId: null,
        layeredRegion: null,
      })
      this.logger.debug(`Volume ${volume.id} deleted successfully (backend=${backend})`)
    } catch (error) {
      this.logger.error(`Error deleting volume ${volume.id}:`, error)
      await this.volumeRepository.save({
        ...volume,
        state: VolumeState.ERROR,
        errorReason: error.message,
      })
    }
  }

  // Delete the volume's layered disk. Always called *before*
  // deleteLayeredBucketPrefix so the disk stops syncing into S3 —
  // otherwise an in-flight write could recreate objects right after we
  // emptied the prefix. Idempotent on missing disks (the control plane
  // returns 404 which is treated as success in `LayeredVolumeClient`).
  private async deleteLayeredDisk(volume: Volume, lockKey: string): Promise<void> {
    if (!volume.layeredDiskId) {
      return
    }
    if (!this.layeredClient.isConfigured()) {
      throw new Error(
        `Volume ${volume.id} has a layered disk but the control plane is not configured on this API. Set LAYERED_API_KEY to retry deletion.`,
      )
    }

    await this.redis.setex(lockKey, 30, '1')
    await this.layeredClient.deleteDisk(
      volume.layeredDiskId,
      volume.layeredRegion || this.layeredClient.getDefaultRegion(),
    )
  }

  // Empties out the volume's prefix from the org's layered bucket. We do
  // NOT delete the bucket itself — it's shared across every layered volume
  // in the organization.
  private async deleteLayeredBucketPrefix(volume: Volume, lockKey: string): Promise<void> {
    if (!this.s3Client) {
      throw new Error(
        `Volume ${volume.id} has layered data but S3 is not configured on this API. Configure S3 to retry deletion.`,
      )
    }
    if (!volume.organizationId) {
      return
    }

    const org = await this.organizationRepository.findOne({ where: { id: volume.organizationId } })
    if (!org?.layeredBucketName) {
      // Nothing to delete: either the org bucket was never provisioned
      // (volume errored before storage was created) or it was already
      // cleared.
      return
    }

    await this.redis.setex(lockKey, 30, '1')
    const prefix = volume.getLayeredBucketPrefix()
    let continuationToken: string | undefined
    do {
      const list = await this.s3Client.send(
        new ListObjectsV2Command({
          Bucket: org.layeredBucketName,
          Prefix: prefix,
          ContinuationToken: continuationToken,
        }),
      )
      if (list.Contents && list.Contents.length > 0) {
        await this.s3Client.send(
          new DeleteObjectsCommand({
            Bucket: org.layeredBucketName,
            Delete: {
              Objects: list.Contents.map((o) => ({ Key: o.Key })),
              Quiet: true,
            },
          }),
        )
      }
      continuationToken = list.NextContinuationToken
    } while (continuationToken)
  }

  // Delete the volume's backing S3 bucket for legacy s3fuse volumes.
  // Tolerates NoSuchBucket so retries after a partial delete are safe.
  private async deletePerVolumeBucket(volume: Volume, lockKey: string): Promise<void> {
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
