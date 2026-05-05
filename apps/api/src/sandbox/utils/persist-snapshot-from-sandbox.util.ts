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
  cpu: number
  gpu: number
  mem: number
  disk: number
  sizeGB?: number
}

/**
 * Inserts a Snapshot row in the ACTIVE state for an image that a runner has
 * already produced (snapshot-from-sandbox), wires up the matching
 * SnapshotRunner record, and emits the CREATED event.
 *
 * Extracted to a free function to avoid a NestJS DI cycle between
 * SnapshotService and JobStateHandlerService: both can call this without
 * importing each other.
 */
export async function persistSnapshotFromSandbox(
  deps: PersistSnapshotFromSandboxDeps,
  params: PersistSnapshotFromSandboxParams,
): Promise<Snapshot> {
  const { snapshotRepository, snapshotRunnerRepository, eventEmitter } = deps

  const size = typeof params.sizeGB === 'number' && Number.isFinite(params.sizeGB) ? params.sizeGB : undefined
  const runnerId = params.runnerId || undefined
  const snapshotId = uuidv4()

  // We should set to active only after a number of snapshot runners had been propagated, leaving as is for now
  const snapshot = snapshotRepository.create({
    id: snapshotId,
    organizationId: params.organizationId,
    name: params.name,
    ref: params.ref,
    state: SnapshotState.ACTIVE,
    cpu: params.cpu,
    gpu: params.gpu,
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
