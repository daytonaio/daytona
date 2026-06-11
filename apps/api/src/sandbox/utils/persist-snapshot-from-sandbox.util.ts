/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException, Logger } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { Repository } from 'typeorm'
import { v4 as uuidv4 } from 'uuid'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { SnapshotEvents } from '../constants/snapshot-events'
import { SnapshotCreatedEvent } from '../events/snapshot-created.event'
import { SnapshotRepository } from '../repositories/snapshot.repository'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { GpuType } from '../enums/gpu-type.enum'

const logger = new Logger('PersistSnapshotFromSandbox')

export interface PersistSnapshotFromSandboxDeps {
  snapshotRepository: SnapshotRepository
  snapshotRunnerRepository: Repository<SnapshotRunner>
  eventEmitter: EventEmitter2
}

export interface PersistSnapshotFromSandboxParams {
  organizationId: string
  name: string
  ref: string
  runnerId?: string | null
  regionId: string
  sandboxClass: SandboxClass
  cpu: number
  gpu: number
  gpuType?: GpuType | null
  mem: number
  disk: number
  sizeGB?: number
}

export type CreatePendingSnapshotFromSandboxParams = Omit<
  PersistSnapshotFromSandboxParams,
  'ref' | 'runnerId' | 'sizeGB'
>

/**
 * Identifies a snapshot-from-sandbox capture record awaiting its result.
 *
 * Capture records are the only PENDING rows with an empty imageName, no
 * buildInfo and no ref: createFromPull rejects empty image names,
 * createFromBuildInfo always sets buildInfo, and activateSnapshotFromSandbox
 * sets ref. The ref check excludes previously-captured snapshots that were
 * deactivated and reactivated (INACTIVE -> PENDING): those keep their empty
 * imageName but carry the ref written at activation, and must go through the
 * regular pull-by-ref flow instead of the capture timeout.
 */
export function isPendingCaptureSnapshot(snapshot: Snapshot): boolean {
  return snapshot.state === SnapshotState.PENDING && !snapshot.buildInfo && !snapshot.imageName && !snapshot.ref
}

/**
 * Inserts the Snapshot record (state=PENDING) when a snapshot-from-sandbox
 * capture is accepted, before the capture is dispatched. A duplicate
 * (organizationId, name) surfaces as ConflictException (HTTP 409) here,
 * before any capture work runs. Emits the CREATED event, which transfers the
 * pending quota counter to current usage — from this point quota settlement
 * follows the record lifecycle (PENDING -> ACTIVE/ERROR state transitions).
 *
 * Extracted to a free function to avoid a NestJS DI cycle between
 * SnapshotService and JobStateHandlerService: both can call this without
 * importing each other.
 */
export async function createPendingSnapshotFromSandbox(
  deps: PersistSnapshotFromSandboxDeps,
  params: CreatePendingSnapshotFromSandboxParams,
): Promise<Snapshot> {
  const { snapshotRepository, eventEmitter } = deps

  const snapshotId = uuidv4()
  const snapshot = snapshotRepository.create({
    id: snapshotId,
    organizationId: params.organizationId,
    name: params.name,
    state: SnapshotState.PENDING,
    sandboxClass: params.sandboxClass,
    cpu: params.cpu,
    gpu: params.gpu,
    gpuType: params.gpuType ?? null,
    mem: params.mem,
    disk: params.disk,
    snapshotRegions: [{ snapshotId, regionId: params.regionId }],
  })

  let inserted: Snapshot
  try {
    inserted = await snapshotRepository.insert(snapshot)
  } catch (error) {
    if ((error as { code?: string }).code === '23505') {
      throw new ConflictException(`Snapshot with name "${params.name}" already exists for this organization`)
    }
    throw error
  }

  eventEmitter.emit(SnapshotEvents.CREATED, new SnapshotCreatedEvent(inserted))
  return inserted
}

/**
 * Marks the capture record ACTIVE once the runner has produced the image,
 * filling in ref/size/initialRunnerId and wiring up the matching
 * SnapshotRunner record. The PENDING -> ACTIVE transition emits STATE_UPDATED
 * via the repository.
 *
 * If no record exists (SNAPSHOT_SANDBOX jobs enqueued before the accept-time
 * insert was deployed), falls back to inserting an ACTIVE record directly. If
 * the record exists but is no longer a pending capture (e.g. already ERROR
 * after a timeout), returns null without touching it.
 */
export async function activateSnapshotFromSandbox(
  deps: PersistSnapshotFromSandboxDeps,
  params: PersistSnapshotFromSandboxParams,
): Promise<Snapshot | null> {
  const { snapshotRepository, snapshotRunnerRepository, eventEmitter } = deps

  const size = typeof params.sizeGB === 'number' && Number.isFinite(params.sizeGB) ? params.sizeGB : undefined
  const runnerId = params.runnerId || undefined

  const existing = await snapshotRepository.findOne({
    where: { organizationId: params.organizationId, name: params.name },
  })

  if (existing) {
    if (!isPendingCaptureSnapshot(existing)) {
      logger.warn(
        `Snapshot "${params.name}" (org ${params.organizationId}) is not a pending capture record (state: ${existing.state}); skipping activation`,
      )
      return null
    }

    const updated = await snapshotRepository.update(existing.id, {
      updateData: {
        state: SnapshotState.ACTIVE,
        ref: params.ref,
        ...(size !== undefined && { size }),
        initialRunnerId: runnerId,
        lastUsedAt: new Date(),
        errorReason: null,
      },
      entity: existing,
    })

    // SnapshotRunner is created only after the Snapshot row is updated so a
    // failure above doesn't leave an orphan SnapshotRunner record pointing at
    // a ref no Snapshot owns.
    if (runnerId) {
      const snapshotRunner = snapshotRunnerRepository.create({
        snapshotRef: params.ref,
        runnerId,
        state: SnapshotRunnerState.READY,
      })
      await snapshotRunnerRepository.save(snapshotRunner)
    }

    return updated
  }

  // Legacy fallback: jobs enqueued before the accept-time record existed have
  // no row to update — insert the ACTIVE record directly.
  const snapshotId = uuidv4()
  const snapshot = snapshotRepository.create({
    id: snapshotId,
    organizationId: params.organizationId,
    name: params.name,
    ref: params.ref,
    state: SnapshotState.ACTIVE,
    sandboxClass: params.sandboxClass,
    cpu: params.cpu,
    gpu: params.gpu,
    gpuType: params.gpuType ?? null,
    mem: params.mem,
    disk: params.disk,
    size,
    initialRunnerId: runnerId,
    lastUsedAt: new Date(),
    snapshotRegions: [{ snapshotId, regionId: params.regionId }],
  })

  let inserted: Snapshot
  try {
    inserted = await snapshotRepository.insert(snapshot)
  } catch (error) {
    if ((error as { code?: string }).code === '23505') {
      throw new ConflictException(`Snapshot with name "${params.name}" already exists for this organization`)
    }
    throw error
  }

  // SnapshotRunner is created only after the Snapshot row is committed so a
  // unique-name conflict above doesn't leave an orphan SnapshotRunner record
  // pointing at a ref no Snapshot owns.
  if (runnerId) {
    const snapshotRunner = snapshotRunnerRepository.create({
      snapshotRef: params.ref,
      runnerId,
      state: SnapshotRunnerState.READY,
    })
    await snapshotRunnerRepository.save(snapshotRunner)
  }

  eventEmitter.emit(SnapshotEvents.CREATED, new SnapshotCreatedEvent(inserted))
  return inserted
}

/**
 * Marks the capture record ERROR when the capture fails. The PENDING -> ERROR
 * transition emits STATE_UPDATED, which decrements current snapshot usage in
 * OrganizationUsageService — the same settlement path pull/build snapshot
 * failures use. No-ops (with a warning) when the record is missing or is not
 * a pending capture, making it idempotent against late or duplicate failure
 * signals.
 */
export async function failSnapshotFromSandbox(
  deps: PersistSnapshotFromSandboxDeps,
  params: { organizationId: string; name: string; errorReason: string },
): Promise<void> {
  const { snapshotRepository } = deps

  const existing = await snapshotRepository.findOne({
    where: { organizationId: params.organizationId, name: params.name },
  })

  if (!existing || !isPendingCaptureSnapshot(existing)) {
    logger.warn(
      `Snapshot "${params.name}" (org ${params.organizationId}) is ${
        existing ? `not a pending capture record (state: ${existing.state})` : 'missing'
      }; skipping failure marking`,
    )
    return
  }

  await snapshotRepository.update(existing.id, {
    updateData: {
      state: SnapshotState.ERROR,
      errorReason: params.errorReason,
    },
    entity: existing,
  })
}
