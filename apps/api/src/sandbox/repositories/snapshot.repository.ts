/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource } from 'typeorm'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotRegion } from '../entities/snapshot-region.entity'
import { Injectable, NotFoundException } from '@nestjs/common'
import { InjectDataSource } from '@nestjs/typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { BaseRepository } from '../../common/repositories/base.repository'
import { SnapshotEvents } from '../constants/snapshot-events'
import { SnapshotStateUpdatedEvent } from '../events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../events/snapshot-removed.event'
import { SnapshotConflictError } from '../errors/snapshot-conflict.error'

@Injectable()
export class SnapshotRepository extends BaseRepository<Snapshot> {
  constructor(@InjectDataSource() dataSource: DataSource, eventEmitter: EventEmitter2) {
    super(dataSource, eventEmitter, Snapshot)
  }

  async insert(snapshot: Snapshot): Promise<Snapshot> {
    const now = new Date()
    if (!snapshot.createdAt) {
      snapshot.createdAt = now
    }
    if (!snapshot.updatedAt) {
      snapshot.updatedAt = now
    }

    await this.dataSource.transaction(async (entityManager) => {
      await entityManager.insert(Snapshot, snapshot)
      if (snapshot.snapshotRegions?.length) {
        await entityManager.insert(SnapshotRegion, snapshot.snapshotRegions)
      }
    })

    return snapshot
  }

  /**
   * @param id - The ID of the snapshot to update.
   * @param params.updateData - The partial data to update.
   *
   * @returns `void` because a raw update is performed.
   */
  async update(id: string, params: { updateData: Partial<Snapshot> }, raw: true): Promise<void>
  /**
   * @param id - The ID of the snapshot to update.
   * @param params.updateData - The partial data to update.
   * @param params.entity - Optional pre-fetched snapshot to use instead of fetching from the database.
   *
   * @returns The updated snapshot.
   */
  async update(id: string, params: { updateData: Partial<Snapshot>; entity?: Snapshot }, raw?: false): Promise<Snapshot>
  async update(
    id: string,
    params: { updateData: Partial<Snapshot>; entity?: Snapshot },
    raw = false,
  ): Promise<Snapshot | void> {
    const { updateData, entity } = params

    if (raw) {
      await this.repository.update(id, updateData)
      return
    }

    const snapshot = entity ?? (await this.findOneBy({ id }))
    if (!snapshot) {
      throw new NotFoundException('Snapshot not found')
    }

    const previousSnapshot = { ...snapshot }

    Object.assign(snapshot, updateData)

    await this.dataSource.transaction(async (entityManager) => {
      const result = await entityManager.update(Snapshot, id, updateData)
      if (!result.affected) {
        throw new SnapshotConflictError()
      }
      snapshot.updatedAt = new Date()
    })

    this.emitUpdateEvents(snapshot, previousSnapshot)

    return snapshot
  }

  async remove(snapshot: Snapshot): Promise<Snapshot> {
    const removed = await super.remove(snapshot)
    this.eventEmitter.emit(SnapshotEvents.REMOVED, new SnapshotRemovedEvent(snapshot))
    return removed
  }

  private emitUpdateEvents(updatedSnapshot: Snapshot, previousSnapshot: Pick<Snapshot, 'state'>): void {
    if (previousSnapshot.state !== updatedSnapshot.state) {
      this.eventEmitter.emit(
        SnapshotEvents.STATE_UPDATED,
        new SnapshotStateUpdatedEvent(updatedSnapshot, previousSnapshot.state, updatedSnapshot.state),
      )
    }
  }
}
