/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject, Logger } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, RemoveEvent, UpdateEvent } from 'typeorm'
import { SnapshotEvents } from '../constants/snapshot-events'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotStateUpdatedEvent } from '../events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../events/snapshot-removed.event'
import { SnapshotLookupCacheInvalidationService } from '../services/snapshot-lookup-cache-invalidation.service'

@EventSubscriber()
export class SnapshotSubscriber implements EntitySubscriberInterface<Snapshot> {
  private readonly logger = new Logger(SnapshotSubscriber.name)

  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  @Inject(SnapshotLookupCacheInvalidationService)
  private snapshotLookupCacheInvalidationService: SnapshotLookupCacheInvalidationService

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Snapshot
  }

  afterInsert(event: InsertEvent<Snapshot>) {
    const entity = event.entity as Snapshot | undefined
    if (!entity) {
      return
    }

    try {
      this.snapshotLookupCacheInvalidationService.invalidate({
        snapshotId: entity.id,
        organizationId: entity.organizationId,
        name: entity.name,
      })
    } catch (error) {
      this.logger.warn(
        `Failed to enqueue snapshot lookup cache invalidation for ${entity.id}: ${error instanceof Error ? error.message : String(error)}`,
      )
    }
  }

  afterUpdate(event: UpdateEvent<Snapshot>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    updatedColumns.forEach((column) => {
      switch (column) {
        case 'state':
          this.eventEmitter.emit(
            SnapshotEvents.STATE_UPDATED,
            new SnapshotStateUpdatedEvent(event.entity as Snapshot, event.databaseEntity[column], event.entity[column]),
          )
          break
        default:
          break
      }
    })

    const entity = event.entity as Snapshot | undefined
    const dbEntity = event.databaseEntity as Snapshot | undefined
    if (!entity || !dbEntity) {
      return
    }

    try {
      this.snapshotLookupCacheInvalidationService.invalidate({
        snapshotId: entity.id,
        organizationId: entity.organizationId,
        name: entity.name,
        previousOrganizationId: dbEntity.organizationId,
        previousName: dbEntity.name,
      })
    } catch (error) {
      this.logger.warn(
        `Failed to enqueue snapshot lookup cache invalidation for ${entity.id}: ${error instanceof Error ? error.message : String(error)}`,
      )
    }
  }

  beforeRemove(event: RemoveEvent<Snapshot>) {
    this.eventEmitter.emit(SnapshotEvents.REMOVED, new SnapshotRemovedEvent(event.databaseEntity as Snapshot))

    const entity = event.databaseEntity as Snapshot | undefined
    if (!entity) {
      return
    }

    try {
      this.snapshotLookupCacheInvalidationService.invalidate({
        snapshotId: entity.id,
        organizationId: entity.organizationId,
        name: entity.name,
      })
    } catch (error) {
      this.logger.warn(
        `Failed to enqueue snapshot lookup cache invalidation for ${entity.id}: ${error instanceof Error ? error.message : String(error)}`,
      )
    }
  }
}
