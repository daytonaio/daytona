/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, RemoveEvent, UpdateEvent } from 'typeorm'
import { SnapshotEvents } from '../constants/snapshot-events'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotCreatedEvent } from '../events/snapshot-created.event'
import { SnapshotStateUpdatedEvent } from '../events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../events/snapshot-removed.event'

@EventSubscriber()
export class SnapshotSubscriber implements EntitySubscriberInterface<Snapshot> {
  constructor(
    dataSource: DataSource,
    @Inject(EventEmitter2)
    private eventEmitter: EventEmitter2,
  ) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Snapshot
  }

  afterInsert(event: InsertEvent<Snapshot>) {
    this.eventEmitter.emit(SnapshotEvents.CREATED, new SnapshotCreatedEvent(event.entity as Snapshot))
  }

  afterUpdate(event: UpdateEvent<Snapshot>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    if (!event.entity) {
      return
    }

    const entity = event.entity as Snapshot

    updatedColumns.forEach((column) => {
      switch (column) {
        case 'state':
          this.eventEmitter.emit(
            SnapshotEvents.STATE_UPDATED,
            new SnapshotStateUpdatedEvent(entity, event.databaseEntity[column], entity[column]),
          )
          break
        default:
          break
      }
    })
  }

  beforeRemove(event: RemoveEvent<Snapshot>) {
    this.eventEmitter.emit(SnapshotEvents.REMOVED, new SnapshotRemovedEvent(event.databaseEntity as Snapshot))
  }
}
