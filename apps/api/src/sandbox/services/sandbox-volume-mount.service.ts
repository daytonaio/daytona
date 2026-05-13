/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { In, Repository } from 'typeorm'
import { SandboxVolumeMount } from '../entities/sandbox-volume.entity'
import { Volume } from '../entities/volume.entity'
import { SandboxVolume } from '../dto/sandbox.dto'
import { EncryptionService } from '../../encryption/encryption.service'
import { LayeredVolumeClient } from './layered/layered-volume.client'
import { VOLUME_BACKEND_LAYERED } from './volume.service'
import { VolumeState } from '../enums/volume-state.enum'

// Handles the per-(sandbox, volume) attachment lifecycle for the layered
// backend: creating `sandbox_volume` rows when a sandbox is created,
// lazily minting + encrypting a layered mount key on first sandbox start,
// and revoking the key + dropping the row on sandbox destroy.
//
// Legacy s3fuse mounts never reach this service — they continue to live
// on the `sandbox.volumes` JSONB column.
@Injectable()
export class SandboxVolumeMountService {
  private readonly logger = new Logger(SandboxVolumeMountService.name)

  constructor(
    @InjectRepository(SandboxVolumeMount)
    private readonly mountRepository: Repository<SandboxVolumeMount>,
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    private readonly encryptionService: EncryptionService,
    private readonly layeredClient: LayeredVolumeClient,
  ) {}

  // Records the layered mounts requested for a new sandbox. Validates that
  // every referenced volume is layered (mixed-backend sandboxes are
  // rejected here, not at the runner) and that each volume is READY.
  //
  // Token minting is deferred to `prepareForStart` so we don't burn a
  // layered control-plane API call on sandboxes that may never start.
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

  // Returns the layered attachments for a sandbox, minting any missing
  // mount keys on demand. Called from `VolumeService.prepareRunnerVolumes`
  // and from `SandboxStartAction` immediately before the runner is asked
  // to start a sandbox.
  //
  // The decrypted token is returned out-of-band (via the second element of
  // the tuple) so the caller can hand it to the runner; the persisted row
  // always carries the encrypted form.
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
        token = await this.layeredClient.mintMountKey({
          diskId: volume.layeredDiskId,
          region: volume.layeredRegion,
          nickname: `sandbox-${sandboxId}`,
        })
        mount.mountKeyEnc = await this.encryptionService.encrypt(token)
        await this.mountRepository.save(mount)
      }

      prepared.push({ mount, volume, mountToken: token })
    }
    return prepared
  }

  // Best-effort revocation of every mount key held by a sandbox, followed
  // by deletion of its `sandbox_volume` rows. Called from the sandbox
  // destroy path. Failures are logged and swallowed: the rows are deleted
  // either way so the next attach starts fresh; orphaned tokens on the
  // control plane are a manual-cleanup concern, not a correctness one.
  async detachVolumesFromSandbox(sandboxId: string): Promise<void> {
    const mounts = await this.mountRepository.find({ where: { sandboxId } })
    if (!mounts.length) {
      return
    }

    const volumes = await this.volumeRepository.find({ where: { id: In(mounts.map((m) => m.volumeId)) } })
    const volumesById = new Map(volumes.map((v) => [v.id, v]))

    for (const mount of mounts) {
      if (!mount.mountKeyEnc) {
        continue
      }
      const volume = volumesById.get(mount.volumeId)
      if (!volume?.layeredDiskId) {
        continue
      }
      try {
        const token = await this.encryptionService.decrypt(mount.mountKeyEnc)
        await this.layeredClient.revokeMountKey(
          volume.layeredDiskId,
          volume.layeredRegion || this.layeredClient.getDefaultRegion(),
          token,
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
