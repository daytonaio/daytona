/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException } from '@nestjs/common'
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

/*
 * Snapshot-from-sandbox lifecycle helpers.
 *
 * A snapshot created from a sandbox gets its Snapshot row inserted in the
 * SNAPSHOTTING state as soon as the operation starts (createSnapshotFromSandboxEntry),
 * is transitioned to ACTIVE once the runner has produced and pushed the image
 * (completeSnapshotFromSandbox), or to ERROR if the operation fails
 * (failSnapshotFromSandbox).
 *
 * Extracted to free functions to avoid a NestJS DI cycle between
 * SnapshotService and JobStateHandlerService: both can call these without
 * importing each other.
 */

export interface SnapshotFromSandboxDeps {
  snapshotRepository: SnapshotRepository
  snapshotRunnerRepository: Repository<SnapshotRunner>
  eventEmitter: EventEmitter2
}

export interface CreateSnapshotFromSandboxEntryParams {
  organizationId: string
  name: string
  regionId: string
  sandboxClass: SandboxClass
  cpu: number
  gpu: number
  gpuType?: GpuType | null
  mem: number
  disk: number
  initialRunnerId?: string
}

export interface CompleteSnapshotFromSandboxParams {
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

export interface FailSnapshotFromSandboxParams {
  organizationId: string
  name: string
  errorReason: string
}

/**
 * Inserts a Snapshot row in the SNAPSHOTTING state before the runner starts
 * producing the image, so the user can track the operation from the start.
 * Emits the CREATED event.
 *
 * @throws {ConflictException} If a snapshot with the same name already exists
 * for the organization.
 */
export async function createSnapshotFromSandboxEntry(
  deps: SnapshotFromSandboxDeps,
  params: CreateSnapshotFromSandboxEntryParams,
): Promise<Snapshot> {
  const { snapshotRepository, eventEmitter } = deps

  const snapshotId = uuidv4()

  const snapshot = snapshotRepository.create({
    id: snapshotId,
    organizationId: params.organizationId,
    name: params.name,
    state: SnapshotState.SNAPSHOTTING,
    sandboxClass: params.sandboxClass,
    cpu: params.cpu,
    gpu: params.gpu,
    gpuType: params.gpuType ?? null,
    mem: params.mem,
    disk: params.disk,
    initialRunnerId: params.initialRunnerId,
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
 * Transitions a SNAPSHOTTING Snapshot row to ACTIVE once the runner has
 * produced the image, filling in the ref/size, and wires up the matching
 * SnapshotRunner record.
 *
 * Returns `null` without touching the row if it exists but is no longer in
 * the SNAPSHOTTING state (e.g. removed or timed out while the runner was
 * working).
 *
 * If no row exists at all, falls back to inserting one directly in the
 * ACTIVE state — this covers operations that were started before the
 * entry-first flow was deployed.
 */
export async function completeSnapshotFromSandbox(
  deps: SnapshotFromSandboxDeps,
  params: CompleteSnapshotFromSandboxParams,
): Promise<Snapshot | null> {
  const { snapshotRepository, snapshotRunnerRepository, eventEmitter } = deps

  const size = typeof params.sizeGB === 'number' && Number.isFinite(params.sizeGB) ? params.sizeGB : undefined
  const runnerId = params.runnerId || undefined

  const existing = await snapshotRepository.findOne({
    where: { organizationId: params.organizationId, name: params.name },
  })

  let snapshot: Snapshot
  if (existing) {
    if (existing.state !== SnapshotState.SNAPSHOTTING) {
      return null
    }

    const updateData: Partial<Snapshot> = {
      state: SnapshotState.ACTIVE,
      ref: params.ref,
      lastUsedAt: new Date(),
    }
    if (size !== undefined) {
      updateData.size = size
    }
    if (runnerId && existing.initialRunnerId !== runnerId) {
      updateData.initialRunnerId = runnerId
    }

    snapshot = await snapshotRepository.update(existing.id, { updateData, entity: existing })
  } else {
    const snapshotId = uuidv4()

    const newSnapshot = snapshotRepository.create({
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

    try {
      snapshot = await snapshotRepository.insert(newSnapshot)
    } catch (error) {
      if ((error as { code?: string }).code === '23505') {
        throw new ConflictException(`Snapshot with name "${params.name}" already exists for this organization`)
      }
      throw error
    }

    eventEmitter.emit(SnapshotEvents.CREATED, new SnapshotCreatedEvent(snapshot))
  }

  // SnapshotRunner is wired up only after the Snapshot row is committed so a
  // failure above doesn't leave an orphan SnapshotRunner record pointing at a
  // ref no Snapshot owns. Snapshotting the same sandbox twice can produce the
  // same content-addressed ref, so reuse an existing record if there is one.
  if (runnerId) {
    const existingSnapshotRunner = await snapshotRunnerRepository.findOne({
      where: { snapshotRef: params.ref, runnerId },
    })
    if (existingSnapshotRunner) {
      if (existingSnapshotRunner.state !== SnapshotRunnerState.READY) {
        existingSnapshotRunner.state = SnapshotRunnerState.READY
        existingSnapshotRunner.errorReason = null
        await snapshotRunnerRepository.save(existingSnapshotRunner)
      }
    } else {
      const snapshotRunner = snapshotRunnerRepository.create({
        snapshotRef: params.ref,
        runnerId,
        state: SnapshotRunnerState.READY,
      })
      await snapshotRunnerRepository.save(snapshotRunner)
    }
  }

  return snapshot
}

/**
 * Transitions a SNAPSHOTTING Snapshot row to ERROR. Returns `null` if the
 * row doesn't exist or is no longer in the SNAPSHOTTING state.
 */
export async function failSnapshotFromSandbox(
  deps: SnapshotFromSandboxDeps,
  params: FailSnapshotFromSandboxParams,
): Promise<Snapshot | null> {
  const { snapshotRepository } = deps

  const existing = await snapshotRepository.findOne({
    where: { organizationId: params.organizationId, name: params.name, state: SnapshotState.SNAPSHOTTING },
  })

  if (!existing) {
    return null
  }

  return snapshotRepository.update(existing.id, {
    updateData: {
      state: SnapshotState.ERROR,
      errorReason: params.errorReason,
    },
    entity: existing,
  })
}
