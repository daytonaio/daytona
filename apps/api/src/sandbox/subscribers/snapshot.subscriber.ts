/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, RemoveEvent } from 'typeorm'
import { SnapshotEvents } from '../constants/snapshot-events'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotCreatedEvent } from '../events/snapshot-created.event'
import { SnapshotRemovedEvent } from '../events/snapshot-removed.event'

@EventSubscriber()
export class SnapshotSubscriber implements EntitySubscriberInterface<Snapshot> {
  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Snapshot
  }

  afterInsert(event: InsertEvent<Snapshot>) {
    this.eventEmitter.emit(SnapshotEvents.CREATED, new SnapshotCreatedEvent(event.entity as Snapshot))
  }

  beforeRemove(event: RemoveEvent<Snapshot>) {
    this.eventEmitter.emit(SnapshotEvents.REMOVED, new SnapshotRemovedEvent(event.databaseEntity as Snapshot))
  }
}
