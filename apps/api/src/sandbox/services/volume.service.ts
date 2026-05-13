/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException, Injectable, Logger, NotFoundException, ServiceUnavailableException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, Not, In } from 'typeorm'
import { Volume } from '../entities/volume.entity'
import { VolumeState } from '../enums/volume-state.enum'
import { CreateVolumeDto } from '../dto/create-volume.dto'
import { v4 as uuidv4 } from 'uuid'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Organization } from '../../organization/entities/organization.entity'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { OrganizationService } from '../../organization/services/organization.service'
import { OrganizationUsageService } from '../../organization/services/organization-usage.service'
import { TypedConfigService } from '../../config/typed-config.service'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { SandboxRepository } from '../repositories/sandbox.repository'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxVolume } from '../dto/sandbox.dto'
import { EncryptionService } from '../../encryption/encryption.service'
import { DtoVolumeDTO } from '@daytona/runner-api-client'
import { VolumeManager } from '../managers/volume.manager'

export const VOLUME_BACKEND_S3FUSE = 's3fuse'
export const VOLUME_BACKEND_EXPERIMENTAL = 'experimental'

export interface PreparedRunnerVolumes {
  volumes: DtoVolumeDTO[]
  // The single backend that all of the sandbox's volumes share. Sandbox start
  // sets `metadata.volumeBackend` to this so the runner picks the matching
  // mounter (host-side s3fuse vs in-container Archil). Undefined when the
  // sandbox has no volumes at all.
  backend?: string
}

@Injectable()
export class VolumeService {
  private readonly logger = new Logger(VolumeService.name)

  constructor(
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    private readonly sandboxRepository: SandboxRepository,
    private readonly organizationService: OrganizationService,
    private readonly organizationUsageService: OrganizationUsageService,
    private readonly configService: TypedConfigService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly encryptionService: EncryptionService,
    private readonly volumeManager: VolumeManager,
  ) {}

  private async validateOrganizationQuotas(
    organization: Organization,
    addedVolumeCount: number,
  ): Promise<{
    pendingVolumeCountIncremented: boolean
  }> {
    // validate usage quotas
    await this.organizationUsageService.incrementPendingVolumeUsage(organization.id, addedVolumeCount)

    const usageOverview = await this.organizationUsageService.getVolumeUsageOverview(organization.id)

    try {
      if (usageOverview.currentVolumeUsage + usageOverview.pendingVolumeUsage > organization.volumeQuota) {
        throw new BadRequestError(`Volume quota exceeded. Maximum allowed: ${organization.volumeQuota}`)
      }
    } catch (error) {
      await this.rollbackPendingUsage(organization.id, addedVolumeCount)
      throw error
    }

    return {
      pendingVolumeCountIncremented: true,
    }
  }

  async rollbackPendingUsage(organizationId: string, pendingVolumeCountIncrement?: number): Promise<void> {
    if (!pendingVolumeCountIncrement) {
      return
    }

    try {
      await this.organizationUsageService.decrementPendingVolumeUsage(organizationId, pendingVolumeCountIncrement)
    } catch (error) {
      this.logger.error(`Error rolling back pending volume usage: ${error}`)
    }
  }

  async create(organization: Organization, createVolumeDto: CreateVolumeDto): Promise<Volume> {
    // The backend is locked at create time so the rest of the volume's
    // lifecycle (provision, mount, delete) has a single source of truth.
    const backend = organization.defaultVolumeBackend || VOLUME_BACKEND_S3FUSE

    // Each backend has its own configuration prerequisite. Fail fast with a
    // clear message rather than letting the async manager get stuck in
    // PENDING_CREATE forever.
    if (backend === VOLUME_BACKEND_S3FUSE && !this.configService.get('s3.endpoint')) {
      throw new ServiceUnavailableException('Object storage is not configured')
    }
    if (backend === VOLUME_BACKEND_EXPERIMENTAL) {
      // The experimental backend stores data in a Daytona-owned S3 bucket
      // and exposes it through an Archil disk that mounts that bucket.
      // Both services therefore have to be configured — there is no
      // archil-managed-storage path.
      if (!this.configService.get('s3.endpoint')) {
        throw new ServiceUnavailableException(
          'Experimental volume backend requires S3 to be configured (the Archil disk is backed by a Daytona-owned S3 bucket). Configure S3 or change the organization default to s3fuse.',
        )
      }
      if (!this.configService.get('archil.apiKey')) {
        throw new ServiceUnavailableException(
          'Experimental volume backend (Archil) is not configured. Set ARCHIL_API_KEY or change the organization default to s3fuse.',
        )
      }
    }

    let pendingVolumeCountIncrement: number | undefined

    try {
      this.organizationService.assertOrganizationIsNotSuspended(organization)

      const newVolumeCount = 1

      const { pendingVolumeCountIncremented } = await this.validateOrganizationQuotas(organization, newVolumeCount)

      if (pendingVolumeCountIncremented) {
        pendingVolumeCountIncrement = newVolumeCount
      }

      const volume = new Volume()

      // Generate ID
      volume.id = uuidv4()

      // Set name from DTO or use ID as default
      volume.name = createVolumeDto.name || volume.id

      // Check if volume with same name already exists for organization
      const existingVolume = await this.volumeRepository.findOne({
        where: {
          organizationId: organization.id,
          name: volume.name,
          state: Not(VolumeState.DELETED),
        },
      })

      if (existingVolume) {
        throw new BadRequestError(`Volume with name ${volume.name} already exists`)
      }

      volume.organizationId = organization.id
      volume.state = VolumeState.PENDING_CREATE
      volume.backend = backend

      const savedVolume = await this.volumeRepository.save(volume)
      this.logger.debug(`Created volume ${savedVolume.id} for organization ${organization.id}`)
      return savedVolume
    } catch (error) {
      await this.rollbackPendingUsage(organization.id, pendingVolumeCountIncrement)
      throw error
    }
  }

  async delete(volumeId: string): Promise<void> {
    const volume = await this.volumeRepository.findOne({
      where: {
        id: volumeId,
      },
    })

    if (!volume) {
      throw new NotFoundException(`Volume with ID ${volumeId} not found`)
    }

    if (volume.state !== VolumeState.READY && volume.state !== VolumeState.ERROR) {
      throw new BadRequestError(
        `Volume must be in '${VolumeState.READY}' or '${VolumeState.ERROR}' state in order to be deleted`,
      )
    }

    // Check if any non-destroyed sandboxes are using this volume
    const sandboxUsingVolume = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .where('sandbox.organizationId = :organizationId', {
        organizationId: volume.organizationId,
      })
      .andWhere('sandbox.volumes @> :volFilter::jsonb', {
        volFilter: JSON.stringify([{ volumeId }]),
      })
      .andWhere('sandbox.desiredState != :destroyed', {
        destroyed: SandboxDesiredState.DESTROYED,
      })
      .select(['sandbox.id', 'sandbox.name'])
      .getOne()

    if (sandboxUsingVolume) {
      throw new ConflictException(
        `Volume cannot be deleted because it is in use by one or more sandboxes (e.g. ${sandboxUsingVolume.name})`,
      )
    }

    // Update state to mark as deleting
    volume.state = VolumeState.PENDING_DELETE
    await this.volumeRepository.save(volume)
    this.logger.debug(`Marked volume ${volumeId} for deletion`)
  }

  // Switches a volume between the s3fuse and experimental backends in
  // place. The volume's S3 bucket (and therefore its data) is preserved;
  // we only attach or detach an Archil disk on top of it.
  //
  // Safety: refuses to switch if any sandbox referencing the volume is
  // currently running (desiredState IN started, resized). A running
  // sandbox is holding a mount via the *previous* backend, and starting a
  // new sandbox after the switch would mount the same bucket through the
  // *new* backend — two writers on the same bucket through different
  // cache layers, which is unsafe. Stop the affected sandboxes and retry.
  //
  // Idempotent: switching to the backend a volume is already on returns
  // the row unchanged.
  async changeBackend(volumeId: string, target: string): Promise<Volume> {
    if (target !== VOLUME_BACKEND_S3FUSE && target !== VOLUME_BACKEND_EXPERIMENTAL) {
      throw new BadRequestError(
        `Invalid volume backend '${target}'. Expected one of: ${VOLUME_BACKEND_S3FUSE}, ${VOLUME_BACKEND_EXPERIMENTAL}.`,
      )
    }

    const volume = await this.findOne(volumeId)
    const currentBackend = volume.backend || VOLUME_BACKEND_S3FUSE

    if (currentBackend === target) {
      return volume
    }

    if (volume.state !== VolumeState.READY) {
      throw new BadRequestError(
        `Volume must be in '${VolumeState.READY}' state to change its backend (currently '${volume.state}').`,
      )
    }

    // Configuration prerequisites for the target backend.
    if (!this.configService.get('s3.endpoint')) {
      throw new ServiceUnavailableException(
        'Object storage is not configured; cannot migrate the volume because both backends rely on its S3 bucket.',
      )
    }
    if (target === VOLUME_BACKEND_EXPERIMENTAL && !this.configService.get('archil.apiKey')) {
      throw new ServiceUnavailableException(
        'Cannot migrate to the experimental backend: Archil is not configured on this API. Set ARCHIL_API_KEY.',
      )
    }

    // Refuse if any sandbox referencing this volume is actively running
    // or transitioning. A `STOPPED` / `ARCHIVED` / `DESTROYED` sandbox
    // doesn't hold a mount, so switching is safe — it'll get the new
    // backend on its next start.
    const blockingSandbox = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .where('sandbox.organizationId = :organizationId', { organizationId: volume.organizationId })
      .andWhere('sandbox.volumes @> :volFilter::jsonb', { volFilter: JSON.stringify([{ volumeId }]) })
      .andWhere('sandbox.desiredState IN (:...active)', {
        active: [SandboxDesiredState.STARTED, SandboxDesiredState.RESIZED],
      })
      .select(['sandbox.id', 'sandbox.name'])
      .getOne()

    if (blockingSandbox) {
      throw new ConflictException(
        `Volume backend cannot be changed because it is currently in use by sandbox '${blockingSandbox.name}'. ` +
          `Stop the sandbox and retry.`,
      )
    }

    // Delegate the resource ops to the manager, then persist the new
    // `backend` value. We update `backend` last so a partial migration
    // (e.g. Archil createDisk failed) leaves the row pointing at the
    // backend whose resources are still intact, and the user's next
    // sandbox start sees consistent state.
    let updated: Volume
    if (target === VOLUME_BACKEND_EXPERIMENTAL) {
      updated = await this.volumeManager.attachArchilDiskTo(volume)
    } else {
      updated = await this.volumeManager.detachArchilDiskFrom(volume)
    }
    updated.backend = target
    const saved = await this.volumeRepository.save(updated)

    this.logger.debug(`Volume ${volume.id} backend changed: ${currentBackend} → ${target}`)
    return saved
  }

  async findOne(volumeId: string): Promise<Volume> {
    const volume = await this.volumeRepository.findOne({
      where: { id: volumeId },
    })

    if (!volume) {
      throw new NotFoundException(`Volume with ID ${volumeId} not found`)
    }

    return volume
  }

  async findAll(organizationId: string, includeDeleted = false): Promise<Volume[]> {
    return this.volumeRepository.find({
      where: {
        organizationId,
        ...(includeDeleted ? {} : { state: Not(VolumeState.DELETED) }),
      },
      order: {
        lastUsedAt: {
          direction: 'DESC',
          nulls: 'LAST',
        },
        createdAt: 'DESC',
      },
    })
  }

  async findByName(organizationId: string, name: string): Promise<Volume> {
    const volume = await this.volumeRepository.findOne({
      where: {
        organizationId,
        name,
        state: Not(VolumeState.DELETED),
      },
    })

    if (!volume) {
      throw new NotFoundException(`Volume with name ${name} not found`)
    }

    return volume
  }

  async validateVolumes(organizationId: string, volumeIdOrNames: string[]): Promise<void> {
    if (!volumeIdOrNames.length) {
      return
    }

    const volumes = await this.volumeRepository.find({
      where: [
        { id: In(volumeIdOrNames), organizationId, state: Not(VolumeState.DELETED) },
        { name: In(volumeIdOrNames), organizationId, state: Not(VolumeState.DELETED) },
      ],
    })

    // Check if all requested volumes were found and are in a READY state
    const foundIds = new Set(volumes.map((v) => v.id))
    const foundNames = new Set(volumes.map((v) => v.name))

    for (const idOrName of volumeIdOrNames) {
      if (!foundIds.has(idOrName) && !foundNames.has(idOrName)) {
        throw new NotFoundException(`Volume '${idOrName}' not found`)
      }
    }

    for (const volume of volumes) {
      if (volume.state !== VolumeState.READY) {
        throw new BadRequestError(`Volume '${volume.name}' is not in a ready state. Current state: ${volume.state}`)
      }
    }
  }

  // Builds the runner-facing volume DTOs for a sandbox by resolving each
  // SandboxVolume reference to its persisted Volume and decorating it with
  // backend-specific fields (e.g. Archil disk + region + decrypted mount
  // token for the experimental backend). Also returns the single shared
  // backend so sandbox-start can stamp it on the sandbox metadata.
  //
  // Mixed backends within a single sandbox are explicitly rejected: the
  // runner picks one mounter per sandbox, and silently dropping volumes is
  // worse than failing fast.
  async prepareRunnerVolumes(sandboxVolumes?: SandboxVolume[]): Promise<PreparedRunnerVolumes> {
    if (!sandboxVolumes?.length) {
      return { volumes: [], backend: undefined }
    }

    const volumeIds = sandboxVolumes.map((v) => v.volumeId)
    const persisted = await this.volumeRepository.find({ where: { id: In(volumeIds) } })
    const persistedById = new Map(persisted.map((v) => [v.id, v]))

    const dtos: DtoVolumeDTO[] = []
    let resolvedBackend: string | undefined

    for (const ref of sandboxVolumes) {
      const volume = persistedById.get(ref.volumeId)
      if (!volume) {
        // Treat as s3fuse to preserve the historical behavior of allowing the
        // runner to fail later with a clearer "volume not found" instead of
        // crashing the start path here. Users immediately hit a NotFound on
        // the sandbox.start call anyway.
        throw new NotFoundException(`Volume ${ref.volumeId} not found`)
      }

      const backend = volume.backend || VOLUME_BACKEND_S3FUSE
      if (resolvedBackend === undefined) {
        resolvedBackend = backend
      } else if (resolvedBackend !== backend) {
        throw new BadRequestError(
          `Sandbox volumes must all share the same backend. Found '${resolvedBackend}' and '${backend}'.`,
        )
      }

      const dto: DtoVolumeDTO = {
        volumeId: ref.volumeId,
        mountPath: ref.mountPath,
        subpath: ref.subpath,
        // Per-mount read-only flag. Honored by both backends:
        //  - s3fuse: enforced via Docker bind mode (`:ro` on the bind spec)
        //    so the host-side mount-s3 can stay shared and writable; only
        //    the in-container view is read-only.
        //  - experimental: passed as `--read-only` to `archil mount` inside
        //    the sandbox. Archil RO mounts don't take a write delegation,
        //    so multiple sandboxes can hold concurrent RO mounts of the
        //    same disk while a separate RW mount is active elsewhere.
        readOnly: ref.readOnly,
      }

      if (backend === VOLUME_BACKEND_EXPERIMENTAL) {
        if (!volume.archilDiskId || !volume.archilRegion || !volume.archilMountTokenEnc) {
          throw new BadRequestError(
            `Volume ${volume.id} uses the experimental backend but is missing Archil provisioning data. ` +
              `It may still be in '${volume.state}' state; wait for it to reach '${VolumeState.READY}' or check the volume's errorReason.`,
          )
        }
        dto.archilDisk = volume.archilDiskId
        dto.archilRegion = volume.archilRegion
        dto.archilMountToken = await this.encryptionService.decrypt(volume.archilMountTokenEnc)
      }

      dtos.push(dto)
    }

    return { volumes: dtos, backend: resolvedBackend }
  }

  async getOrganizationId(params: { id: string } | { name: string; organizationId: string }): Promise<string> {
    if ('id' in params) {
      const volume = await this.volumeRepository.findOneOrFail({
        where: {
          id: params.id,
        },
        select: ['organizationId'],
        loadEagerRelations: false,
      })
      return volume.organizationId
    }

    const volume = await this.volumeRepository.findOneOrFail({
      where: {
        name: params.name,
        organizationId: params.organizationId,
      },
      select: ['organizationId'],
      loadEagerRelations: false,
    })

    return volume.organizationId
  }

  @OnEvent(SandboxEvents.CREATED)
  private async handleSandboxCreatedEvent(event: SandboxCreatedEvent) {
    if (!event.sandbox.volumes.length) {
      return
    }

    try {
      const volumeIds = event.sandbox.volumes.map((vol) => vol.volumeId)
      const volumes = await this.volumeRepository.find({ where: { id: In(volumeIds) } })

      const results = await Promise.allSettled(
        volumes.map(async (volume) => {
          // Update once per minute at most
          if (!(await this.redisLockProvider.lock(`volume:${volume.id}:update-last-used`, 60))) {
            return
          }
          volume.lastUsedAt = event.sandbox.createdAt
          return this.volumeRepository.save(volume)
        }),
      )

      results.forEach((result) => {
        if (result.status === 'rejected') {
          this.logger.error(
            `Failed to update volume lastUsedAt timestamp for sandbox ${event.sandbox.id}: ${result.reason}`,
          )
        }
      })
    } catch (err) {
      this.logger.error(err)
    }
  }
}
