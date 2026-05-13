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
import { DtoVolumeDTO } from '@daytona/runner-api-client'
import { SandboxVolumeMountService } from './sandbox-volume-mount.service'

export const VOLUME_BACKEND_S3FUSE = 's3fuse'
export const VOLUME_BACKEND_LAYERED = 'layered'

export interface PreparedRunnerVolumes {
  volumes: DtoVolumeDTO[]
  // The single backend that all of the sandbox's volumes share. Sandbox start
  // sets `metadata.volumeBackend` to this so the runner picks the matching
  // mounter (host-side s3fuse vs in-container layered). Undefined when the
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
    private readonly sandboxVolumeMountService: SandboxVolumeMountService,
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
    if (backend === VOLUME_BACKEND_LAYERED) {
      // The layered backend stores data in a per-organization S3 bucket
      // and exposes each volume through a control-plane disk that mounts
      // a `<volumeId>/` prefix of that bucket. Both services have to be
      // configured.
      if (!this.configService.get('s3.endpoint')) {
        throw new ServiceUnavailableException(
          'Layered volume backend requires S3 to be configured (the layered disk is backed by a Daytona-owned S3 bucket). Configure S3 or change the organization default to s3fuse.',
        )
      }
      if (!this.configService.get('layered.apiKey')) {
        throw new ServiceUnavailableException(
          'Layered volume backend is not configured. Set LAYERED_API_KEY or change the organization default to s3fuse.',
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
      this.logger.debug(`Created volume ${savedVolume.id} for organization ${organization.id} (backend=${backend})`)
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

    // Refuse if any non-destroyed sandbox is using this volume. For layered
    // volumes we check `sandbox_volume`; for s3fuse the reference lives in
    // the sandbox JSONB column.
    if (volume.backend === VOLUME_BACKEND_LAYERED) {
      const activeMounts = await this.sandboxVolumeMountService.findAllForVolume(volumeId)
      if (activeMounts.length) {
        const activeIds = activeMounts.map((m) => m.sandboxId)
        const stillRunning = await this.sandboxRepository
          .createQueryBuilder('sandbox')
          .where('sandbox.id IN (:...ids)', { ids: activeIds })
          .andWhere('sandbox.desiredState != :destroyed', { destroyed: SandboxDesiredState.DESTROYED })
          .select(['sandbox.id', 'sandbox.name'])
          .getOne()
        if (stillRunning) {
          throw new ConflictException(
            `Volume cannot be deleted because it is in use by one or more sandboxes (e.g. ${stillRunning.name})`,
          )
        }
      }
    } else {
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
    }

    // Update state to mark as deleting
    volume.state = VolumeState.PENDING_DELETE
    await this.volumeRepository.save(volume)
    this.logger.debug(`Marked volume ${volumeId} for deletion`)
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

  // Resolves the per-sandbox volume DTOs that the runner expects, branching
  // on storage layout:
  //
  //  - sandboxes with rows in `sandbox_volume` (layered backend) — token is
  //    minted on demand by `SandboxVolumeMountService.prepareForStart`.
  //  - sandboxes with entries in `sandbox.volumes` JSONB (legacy s3fuse) —
  //    DTOs are built straight from the JSONB references.
  //
  // The two paths are mutually exclusive by construction: layered volumes
  // are never written to the JSONB column, and s3fuse volumes are never
  // written to `sandbox_volume`. Sandboxes pre-dating the layered backend
  // therefore continue to work bit-for-bit on the legacy path.
  async prepareRunnerVolumes(sandboxId: string, sandboxVolumesJsonb?: SandboxVolume[]): Promise<PreparedRunnerVolumes> {
    const layered = await this.sandboxVolumeMountService.prepareForStart(sandboxId)
    if (layered.length > 0) {
      if (sandboxVolumesJsonb?.length) {
        // Should be impossible (we never persist both), but treat as a
        // hard error so we surface the inconsistency instead of silently
        // dropping mounts.
        throw new BadRequestError(
          `Sandbox ${sandboxId} has both legacy s3fuse volume references and layered sandbox_volume rows. ` +
            `Refusing to start; one of the two must be cleaned up.`,
        )
      }
      const dtos: DtoVolumeDTO[] = layered.map(({ mount, volume, mountToken }) => ({
        volumeId: mount.volumeId,
        mountPath: mount.mountPath,
        subpath: mount.subpath ?? undefined,
        readOnly: mount.readOnly,
        layeredDisk: volume.layeredDiskId,
        layeredRegion: volume.layeredRegion,
        layeredMountToken: mountToken,
      }))
      return { volumes: dtos, backend: VOLUME_BACKEND_LAYERED }
    }

    if (!sandboxVolumesJsonb?.length) {
      return { volumes: [], backend: undefined }
    }

    const volumeIds = sandboxVolumesJsonb.map((v) => v.volumeId)
    const persisted = await this.volumeRepository.find({ where: { id: In(volumeIds) } })
    const persistedById = new Map(persisted.map((v) => [v.id, v]))

    const dtos: DtoVolumeDTO[] = []
    for (const ref of sandboxVolumesJsonb) {
      const volume = persistedById.get(ref.volumeId)
      if (!volume) {
        throw new NotFoundException(`Volume ${ref.volumeId} not found`)
      }
      if ((volume.backend || VOLUME_BACKEND_S3FUSE) !== VOLUME_BACKEND_S3FUSE) {
        // Defense-in-depth: the create path forbids putting layered
        // volumes on the JSONB column, but if somehow one ended up here
        // the runner would silently try to s3fuse-mount it. Fail loudly.
        throw new BadRequestError(
          `Volume ${volume.id} uses backend '${volume.backend}' but is referenced via the legacy s3fuse JSONB column.`,
        )
      }
      dtos.push({
        volumeId: ref.volumeId,
        mountPath: ref.mountPath,
        subpath: ref.subpath,
        // Per-mount read-only flag honored by both backends. The s3fuse
        // path enforces it via Docker bind mode (`:ro` on the bind spec)
        // so the host-side mount-s3 can stay shared and writable; only
        // the in-container view is read-only.
        readOnly: ref.readOnly,
      })
    }

    return { volumes: dtos, backend: VOLUME_BACKEND_S3FUSE }
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
