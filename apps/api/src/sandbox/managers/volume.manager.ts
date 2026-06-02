/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject, Injectable, Logger, OnApplicationBootstrap, OnApplicationShutdown, OnModuleInit } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, In, IsNull } from 'typeorm'
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
import { LAYERED_VOLUME_PROVIDER, LayeredVolumeProvider, DiskMount } from '../services/layered/layered-volume.provider'
import { VOLUME_BACKEND_LAYERED, VOLUME_BACKEND_S3FUSE } from '../services/volume.service'
import { Organization } from '../../organization/entities/organization.entity'
import { EncryptionService } from '../../encryption/encryption.service'
import { Region } from '../../region/entities/region.entity'
import { awsRegionFromStorageRegion, layeredBucketNameFor } from '../services/layered/layered-bucket-name'

const VOLUME_STATE_LOCK_KEY = 'volume-state-'

@Injectable()
export class VolumeManager
  implements OnModuleInit, TrackableJobExecutions, OnApplicationShutdown, OnApplicationBootstrap
{
  activeJobs = new Set<string>()

  private readonly logger = new Logger(VolumeManager.name)
  private processingVolumes: Set<string> = new Set()
  private skipTestConnection = false
  // default client; used by s3fuse and legacy layered
  private s3Client: S3Client | null = null
  // region-pinned clients for layered volumes, keyed by AWS slug; same creds, only `region` differs
  private regionScopedS3Clients: Map<string, S3Client> = new Map()

  constructor(
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    @InjectRepository(Organization)
    private readonly organizationRepository: Repository<Organization>,
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
    private readonly configService: TypedConfigService,
    @InjectRedis() private readonly redis: Redis,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly schedulerRegistry: SchedulerRegistry,
    @Inject(LAYERED_VOLUME_PROVIDER) private readonly layeredClient: LayeredVolumeProvider,
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

  // lazily builds/caches a client scoped to `awsRegion`; null when S3 is unconfigured
  private getS3ClientForAwsRegion(awsRegion: string): S3Client | null {
    if (!this.configService.get('s3.endpoint')) {
      return null
    }
    const cached = this.regionScopedS3Clients.get(awsRegion)
    if (cached) {
      return cached
    }

    const endpoint = this.configService.getOrThrow('s3.endpoint')
    const accessKeyId = this.configService.getOrThrow('s3.accessKey')
    const secretAccessKey = this.configService.getOrThrow('s3.secretKey')

    const client = new S3Client({
      endpoint: this.resolveRegionalS3Endpoint(endpoint, awsRegion),
      region: awsRegion,
      credentials: { accessKeyId, secretAccessKey },
      forcePathStyle: true,
    })
    this.regionScopedS3Clients.set(awsRegion, client)
    return client
  }

  // AWS rejects SigV4 requests whose region doesn't match the host, so route AWS to the
  // per-region host. S3-compatible providers (MinIO, R2) aren't region-routable: leave as-is.
  private resolveRegionalS3Endpoint(configuredEndpoint: string, awsRegion: string): string {
    const withProtocol = configuredEndpoint.startsWith('http') ? configuredEndpoint : `http://${configuredEndpoint}`
    let hostname: string
    try {
      hostname = new URL(withProtocol).hostname
    } catch {
      return withProtocol
    }
    if (!/(^|\.)amazonaws\.com$/i.test(hostname)) {
      return withProtocol
    }
    return `https://s3.${awsRegion}.amazonaws.com`
  }

  // bucketRegion is null for legacy volumes (regionId IS NULL) to preserve the historical
  // bucket name + default client; otherwise it's Region.storageRegion ?? env fallback.
  private async resolveStorageRegionForLayeredVolume(volume: Volume): Promise<{
    bucketRegion: string | null
    effectiveStorageRegion: string
  }> {
    const fallback = this.configService.get('layered.defaultStorageRegion')

    // legacy volumes rely on the env fallback, so it must be set
    if (!volume.regionId) {
      if (!fallback) {
        throw new Error(
          `Cannot resolve storage region for layered volume ${volume.id}: LAYERED_DEFAULT_STORAGE_REGION is unset.`,
        )
      }
      return { bucketRegion: null, effectiveStorageRegion: fallback }
    }

    const region = await this.regionRepository.findOne({ where: { id: volume.regionId } })
    const storageRegion = region?.storageRegion || fallback
    if (!storageRegion) {
      throw new Error(
        `Cannot resolve storage region for layered volume ${volume.id}: region '${volume.regionId}' has no storageRegion and LAYERED_DEFAULT_STORAGE_REGION is unset.`,
      )
    }
    return { bucketRegion: storageRegion, effectiveStorageRegion: storageRegion }
  }

  // run the cron if at least one backend is configured; per-volume branching routes to the
  // matching backend. OR'd (not AND'd) so the cron still runs for s3fuse-only volumes.
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
    this.schedulerRegistry.getCronJob('meter-volume-storage').start()
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

    await this.redis.setex(lockKey, 30, '1')

    await this.volumeRepository.save({
      ...volume,
      state: VolumeState.READY,
    })
  }

  // creates + tags the per-volume s3fuse bucket; idempotent on retry
  // (BucketAlreadyOwnedByYou is treated as success).
  private async createAndTagPerVolumeS3Bucket(volume: Volume, lockKey: string): Promise<void> {
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

  // idempotently provisions the layered bucket. branches: BYOB → user bucket;
  // region-pinned → `dt-vl-<orgId>-<region>` on the per-region client;
  // legacy (regionId NULL) → `daytona-volume-layered-<orgId>` on the default client.
  private async ensureLayeredBucket(volume: Volume): Promise<{
    bucketName: string
    org: Organization
    storageRegion: string
    s3Client: S3Client
  }> {
    if (!volume.organizationId) {
      throw new Error(`Volume ${volume.id} has no organizationId; cannot provision a per-organization layered bucket.`)
    }

    const org = await this.organizationRepository.findOne({ where: { id: volume.organizationId } })
    if (!org) {
      throw new Error(`Organization ${volume.organizationId} not found while provisioning layered volume ${volume.id}`)
    }

    if (org.customBucketConfig) {
      // BYOB skips region pinning; storageRegion is just a placeholder for the control plane call
      return {
        bucketName: org.customBucketConfig.bucketName,
        org,
        storageRegion: this.configService.get('layered.defaultStorageRegion'),
        s3Client: this.s3Client as S3Client,
      }
    }

    const { bucketRegion, effectiveStorageRegion } = await this.resolveStorageRegionForLayeredVolume(volume)
    const bucketName = layeredBucketNameFor(org.id, bucketRegion)
    const awsRegion = awsRegionFromStorageRegion(effectiveStorageRegion)
    const s3Client = bucketRegion === null ? this.s3Client : this.getS3ClientForAwsRegion(awsRegion)

    if (!s3Client) {
      throw new Error(
        `Volume ${volume.id} requires layered storage but S3 is not configured on this API. The layered backend stores volume data in a Daytona-owned S3 bucket and requires S3 to be configured.`,
      )
    }

    try {
      await s3Client.send(
        new CreateBucketCommand({
          Bucket: bucketName,
          // AWS rejects LocationConstraint=us-east-1 (default region).
          ...(bucketRegion !== null && awsRegion !== 'us-east-1'
            ? { CreateBucketConfiguration: { LocationConstraint: awsRegion as never } }
            : {}),
        }),
      )
    } catch (error) {
      if (error?.name !== 'BucketAlreadyOwnedByYou') {
        throw error
      }
    }

    await s3Client.send(
      new PutBucketTaggingCommand({
        Bucket: bucketName,
        Tagging: {
          TagSet: [
            { Key: 'OrganizationId', Value: org.id },
            { Key: 'Purpose', Value: 'layered-volumes' },
            { Key: 'Environment', Value: this.configService.get('environment') },
            ...(bucketRegion !== null ? [{ Key: 'StorageRegion', Value: bucketRegion }] : []),
          ],
        },
      }),
    )

    // only legacy bucket names are persisted; per-region names are derived deterministically
    if (bucketRegion === null && org.layeredBucketName !== bucketName) {
      org.layeredBucketName = bucketName
      await this.organizationRepository.save(org)
    }

    return { bucketName, org, storageRegion: effectiveStorageRegion, s3Client }
  }

  // builds the mount config backing the disk with the org bucket + the volume's prefix.
  // BYOB uses the org's own credentials; otherwise falls back to platform-wide S3 config.
  private async buildLayeredMount(bucketName: string, volume: Volume, org: Organization): Promise<DiskMount> {
    let accessKeyId: string
    let secretAccessKey: string
    let endpoint: string | undefined

    if (org.customBucketConfig) {
      accessKeyId = await this.encryptionService.decrypt(org.customBucketConfig.accessKeyIdEnc)
      secretAccessKey = await this.encryptionService.decrypt(org.customBucketConfig.secretAccessKeyEnc)
      endpoint = org.customBucketConfig.endpoint
    } else {
      endpoint = this.configService.getOrThrow('s3.endpoint') as string
      accessKeyId = this.configService.getOrThrow('s3.accessKey') as string
      secretAccessKey = this.configService.getOrThrow('s3.secretKey') as string
    }

    if (endpoint) {
      const endpointWithProtocol = endpoint.startsWith('http') ? endpoint : `https://${endpoint}`

      let isAws = false
      try {
        isAws = /(^|\.)amazonaws\.com$/i.test(new URL(endpointWithProtocol).hostname)
      } catch {
        // malformed endpoint — fall through to s3-compatible
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

    // no endpoint means native AWS S3 (BYOB user omitted endpoint)
    return {
      type: 's3',
      bucketName,
      bucketPrefix: volume.getLayeredBucketPrefix(),
      accessKeyId,
      secretAccessKey,
    }
  }

  private async provisionLayeredDisk(volume: Volume, lockKey: string): Promise<void> {
    await this.assertLayeredProvisioningPossible(volume)

    await this.redis.setex(lockKey, 30, '1')
    const { bucketName, org, storageRegion } = await this.ensureLayeredBucket(volume)

    // create the disk mounting the bucket at this volume's prefix. the org bucket is not
    // rolled back on failure — it's shared and the next volume create reuses it.
    await this.redis.setex(lockKey, 30, '1')
    let updated: Volume
    try {
      updated = await this.ensureLayeredDiskFor(volume, bucketName, org, storageRegion)
    } catch (error) {
      this.logger.error(`Layered createDisk failed for volume ${volume.id}`, error)
      throw error
    }

    await this.redis.setex(lockKey, 30, '1')
    await this.volumeRepository.save({ ...updated, state: VolumeState.READY })
  }

  // idempotent: an existing `layeredDiskId` means the disk exists, so return as-is (retry-safe).
  // createDisk's initial token is one-time and discarded; per-sandbox tokens are minted on
  // first attach via `LayeredVolumeClient.mintMountKey`.
  private async ensureLayeredDiskFor(
    volume: Volume,
    bucketName: string,
    org: Organization,
    storageRegion: string,
  ): Promise<Volume> {
    if (volume.layeredDiskId) {
      return volume
    }

    const created = await this.layeredClient.createDisk({
      name: volume.getLayeredDiskName(),
      region: storageRegion,
      mount: await this.buildLayeredMount(bucketName, volume, org),
    })

    return await this.volumeRepository.save({
      ...volume,
      layeredDiskId: created.diskId,
      layeredRegion: created.region,
    })
  }

  private async assertLayeredProvisioningPossible(volume: Volume): Promise<void> {
    if (!this.layeredClient.isConfigured()) {
      throw new Error(
        `Volume ${volume.id} requires the layered backend but its control plane is not configured on this API. Set LAYERED_API_KEY.`,
      )
    }

    // S3 is only required when the org doesn't have a custom bucket (BYOB).
    if (!this.s3Client) {
      const org = await this.organizationRepository.findOne({ where: { id: volume.organizationId } })
      if (!org?.customBucketConfig) {
        throw new Error(
          `Volume ${volume.id} requires the layered backend but S3 is not configured on this API. Either configure S3 or set a custom bucket on the organization.`,
        )
      }
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

  // must run before deleteLayeredBucketPrefix so the disk stops syncing into S3, else an
  // in-flight write could recreate objects we just emptied. idempotent on missing disks (404).
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

  // empties the volume's prefix from the org's layered bucket; does NOT delete the bucket
  // itself, which is shared across every layered volume in the (org, storageRegion) pair.
  private async deleteLayeredBucketPrefix(volume: Volume, lockKey: string): Promise<void> {
    if (!volume.organizationId) {
      return
    }

    const org = await this.organizationRepository.findOne({ where: { id: volume.organizationId } })

    let bucketName: string | undefined
    let client: S3Client | null = null

    if (org?.customBucketConfig) {
      bucketName = org.customBucketConfig.bucketName
      const accessKeyId = await this.encryptionService.decrypt(org.customBucketConfig.accessKeyIdEnc)
      const secretAccessKey = await this.encryptionService.decrypt(org.customBucketConfig.secretAccessKeyEnc)
      const endpoint = org.customBucketConfig.endpoint
      client = new S3Client({
        ...(endpoint && { endpoint: endpoint.startsWith('http') ? endpoint : `https://${endpoint}` }),
        ...(org.customBucketConfig.region && { region: org.customBucketConfig.region }),
        credentials: { accessKeyId, secretAccessKey },
        forcePathStyle: true,
      })
    } else if (volume.regionId) {
      const { bucketRegion, effectiveStorageRegion } = await this.resolveStorageRegionForLayeredVolume(volume)
      bucketName = layeredBucketNameFor(org!.id, bucketRegion)
      client = this.getS3ClientForAwsRegion(awsRegionFromStorageRegion(effectiveStorageRegion))
    } else {
      bucketName = org?.layeredBucketName ?? undefined
      client = this.s3Client
    }

    if (!bucketName) {
      return
    }
    if (!client) {
      throw new Error(
        `Volume ${volume.id} has layered data but S3 is not configured on this API. Configure S3 to retry deletion.`,
      )
    }

    await this.redis.setex(lockKey, 30, '1')
    const prefix = volume.getLayeredBucketPrefix()
    let continuationToken: string | undefined
    do {
      const list = await client.send(
        new ListObjectsV2Command({
          Bucket: bucketName,
          Prefix: prefix,
          ContinuationToken: continuationToken,
        }),
      )
      if (list.Contents && list.Contents.length > 0) {
        await client.send(
          new DeleteObjectsCommand({
            Bucket: bucketName,
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

  // ──────────────────────────────────────────────────────────────────────
  // Volume storage metering
  // ──────────────────────────────────────────────────────────────────────

  @Cron(CronExpression.EVERY_MINUTE, { name: 'meter-volume-storage', waitForCompletion: true, disabled: true })
  @TrackJobExecution()
  @LogExecution('meter-volume-storage')
  @WithInstrumentation()
  async meterVolumeStorage() {
    if (!this.anyBackendConfigured) {
      return
    }

    const lockKey = 'meter-volume-storage'
    if (!(await this.redisLockProvider.lock(lockKey, 55))) {
      return
    }

    try {
      // oldest-checked-first so every volume gets measured eventually
      const volumes = await this.volumeRepository
        .createQueryBuilder('volume')
        .where('volume.state = :state', { state: VolumeState.READY })
        .orderBy('volume.storageCheckedAt', 'ASC', 'NULLS FIRST')
        .take(100)
        .getMany()

      for (const volume of volumes) {
        try {
          const sizeMb = await this.measureVolumeSizeMb(volume)
          if (sizeMb !== null) {
            await this.volumeRepository.update(volume.id, {
              currentStorageMb: sizeMb,
              storageCheckedAt: new Date(),
            })
          }
        } catch (error) {
          this.logger.warn(`Failed to measure storage for volume ${volume.id}: ${error?.message ?? error}`)
        }
      }
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private async measureVolumeSizeMb(volume: Volume): Promise<number | null> {
    const backend = volume.backend || VOLUME_BACKEND_S3FUSE
    let client: S3Client | null = null
    let bucketName: string | undefined
    let prefix: string | undefined

    if (backend === VOLUME_BACKEND_LAYERED) {
      const org = await this.organizationRepository.findOne({ where: { id: volume.organizationId } })
      if (org?.customBucketConfig) {
        const accessKeyId = await this.encryptionService.decrypt(org.customBucketConfig.accessKeyIdEnc)
        const secretAccessKey = await this.encryptionService.decrypt(org.customBucketConfig.secretAccessKeyEnc)
        const endpoint = org.customBucketConfig.endpoint
        client = new S3Client({
          ...(endpoint && { endpoint: endpoint.startsWith('http') ? endpoint : `https://${endpoint}` }),
          ...(org.customBucketConfig.region && { region: org.customBucketConfig.region }),
          credentials: { accessKeyId, secretAccessKey },
          forcePathStyle: true,
        })
        bucketName = org.customBucketConfig.bucketName
      } else if (volume.regionId && org) {
        const { bucketRegion, effectiveStorageRegion } = await this.resolveStorageRegionForLayeredVolume(volume)
        bucketName = layeredBucketNameFor(org.id, bucketRegion)
        client = this.getS3ClientForAwsRegion(awsRegionFromStorageRegion(effectiveStorageRegion))
      } else {
        client = this.s3Client
        bucketName = org?.layeredBucketName ?? undefined
      }
      prefix = volume.getLayeredBucketPrefix()
    } else {
      client = this.s3Client
      bucketName = volume.getBucketName()
      prefix = undefined
    }

    if (!client || !bucketName) {
      return null
    }

    let totalBytes = 0
    let continuationToken: string | undefined
    do {
      const list = await client.send(
        new ListObjectsV2Command({
          Bucket: bucketName,
          Prefix: prefix,
          ContinuationToken: continuationToken,
        }),
      )
      if (list.Contents) {
        for (const obj of list.Contents) {
          totalBytes += obj.Size ?? 0
        }
      }
      continuationToken = list.NextContinuationToken
    } while (continuationToken)

    return totalBytes / (1024 * 1024)
  }

  // deletes the legacy s3fuse bucket; tolerates NoSuchBucket so partial-delete retries are safe
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
