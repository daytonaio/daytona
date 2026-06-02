/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Inject, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { In, Repository } from 'typeorm'
import { SandboxVolumeMount } from '../entities/sandbox-volume.entity'
import { Volume } from '../entities/volume.entity'
import { SandboxVolume } from '../dto/sandbox.dto'
import { EncryptionService } from '../../encryption/encryption.service'
import { LAYERED_VOLUME_PROVIDER, LayeredVolumeProvider } from './layered/layered-volume.provider'
import { VOLUME_BACKEND_LAYERED } from './volume.service'
import { VolumeState } from '../enums/volume-state.enum'

// Per-(sandbox, volume) attachment lifecycle for the layered backend: create
// `sandbox_volume` rows on sandbox create, lazily mint + encrypt a mount key on
// first start, revoke + drop the row on destroy. Legacy s3fuse mounts never reach
// this service; they stay on the `sandbox.volumes` JSONB column.
@Injectable()
export class SandboxVolumeMountService {
  private readonly logger = new Logger(SandboxVolumeMountService.name)

  constructor(
    @InjectRepository(SandboxVolumeMount)
    private readonly mountRepository: Repository<SandboxVolumeMount>,
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    private readonly encryptionService: EncryptionService,
    @Inject(LAYERED_VOLUME_PROVIDER) private readonly layeredClient: LayeredVolumeProvider,
  ) {}

  // Records requested layered mounts for a new sandbox, rejecting non-layered or
  // not-READY volumes here rather than at the runner. Token minting is deferred to
  // `prepareForStart` to avoid a control-plane call for sandboxes that never start.
  async attachVolumesToSandbox(
    sandboxId: string,
    refs: SandboxVolume[] | undefined,
    volumesById: Map<string, Volume>,
  ): Promise<SandboxVolumeMount[]> {
    if (!refs?.length) {
      return []
    }

    const rows: SandboxVolumeMount[] = []
    for (const ref of refs) {
      const volume = volumesById.get(ref.volumeId)
      if (!volume) {
        throw new NotFoundException(`Volume ${ref.volumeId} not found`)
      }
      if ((volume.backend || '') !== VOLUME_BACKEND_LAYERED) {
        throw new BadRequestException(
          `Volume ${volume.id} uses backend '${volume.backend}' which is not supported by the layered mount path`,
        )
      }
      if (volume.state !== VolumeState.READY) {
        throw new BadRequestException(
          `Volume ${volume.id} is not ready (state=${volume.state}); cannot attach to sandbox ${sandboxId}.`,
        )
      }

      const row = this.mountRepository.create({
        sandboxId,
        volumeId: ref.volumeId,
        mountPath: ref.mountPath,
        subpath: ref.subpath ?? null,
        readOnly: Boolean(ref.readOnly),
        mountKeyEnc: null,
      })
      rows.push(await this.mountRepository.save(row))
    }
    return rows
  }

  // Returns a sandbox's layered attachments, minting missing mount keys on demand.
  // The decrypted token is returned for the caller to hand to the runner; the
  // persisted row always carries only the encrypted form.
  async prepareForStart(
    sandboxId: string,
  ): Promise<Array<{ mount: SandboxVolumeMount; volume: Volume; mountToken: string }>> {
    const mounts = await this.mountRepository.find({ where: { sandboxId } })
    if (!mounts.length) {
      return []
    }

    const volumes = await this.volumeRepository.find({ where: { id: In(mounts.map((m) => m.volumeId)) } })
    const volumesById = new Map(volumes.map((v) => [v.id, v]))

    const prepared: Array<{ mount: SandboxVolumeMount; volume: Volume; mountToken: string }> = []
    for (const mount of mounts) {
      const volume = volumesById.get(mount.volumeId)
      if (!volume) {
        throw new NotFoundException(`Volume ${mount.volumeId} not found for sandbox_volume row ${mount.id}`)
      }
      if (!volume.layeredDiskId || !volume.layeredRegion) {
        throw new BadRequestException(
          `Volume ${volume.id} is missing layered disk metadata; wait for it to reach state=ready or check errorReason.`,
        )
      }

      let token: string
      if (mount.mountKeyEnc) {
        token = await this.encryptionService.decrypt(mount.mountKeyEnc)
      } else {
        const result = await this.layeredClient.mintMountKey({
          diskId: volume.layeredDiskId,
          region: volume.layeredRegion,
          nickname: `sandbox-${sandboxId}`,
        })
        token = result.token
        mount.mountKeyEnc = await this.encryptionService.encrypt(result.token)
        mount.mountIdentifierEnc = await this.encryptionService.encrypt(result.identifier)
        await this.mountRepository.save(mount)
      }

      prepared.push({ mount, volume, mountToken: token })
    }
    return prepared
  }

  // Best-effort revocation of a sandbox's mount keys, then deletion of its
  // `sandbox_volume` rows. Revoke failures are logged and swallowed; rows are
  // dropped regardless, leaving orphaned control-plane tokens as manual cleanup.
  async detachVolumesFromSandbox(sandboxId: string): Promise<void> {
    const mounts = await this.mountRepository.find({ where: { sandboxId } })
    if (!mounts.length) {
      return
    }

    const volumes = await this.volumeRepository.find({ where: { id: In(mounts.map((m) => m.volumeId)) } })
    const volumesById = new Map(volumes.map((v) => [v.id, v]))

    for (const mount of mounts) {
      if (!mount.mountIdentifierEnc) {
        continue
      }
      const volume = volumesById.get(mount.volumeId)
      if (!volume?.layeredDiskId) {
        continue
      }
      try {
        const identifier = await this.encryptionService.decrypt(mount.mountIdentifierEnc)
        await this.layeredClient.revokeMountKey(
          volume.layeredDiskId,
          volume.layeredRegion || this.layeredClient.getDefaultRegion(),
          identifier,
        )
      } catch (error) {
        this.logger.warn(
          `Failed to revoke layered mount key for sandbox=${sandboxId} volume=${mount.volumeId}: ${error?.message ?? error}`,
        )
      }
    }

    await this.mountRepository.delete({ sandboxId })
  }

  async findAllForSandbox(sandboxId: string): Promise<SandboxVolumeMount[]> {
    return this.mountRepository.find({ where: { sandboxId } })
  }

  async findAllForVolume(volumeId: string): Promise<SandboxVolumeMount[]> {
    return this.mountRepository.find({ where: { volumeId } })
  }
}
