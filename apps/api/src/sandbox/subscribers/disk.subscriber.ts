/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, UpdateEvent } from 'typeorm'
import { DiskEvents } from '../constants/disk-events'
import { Disk } from '../entities/disk.entity'
import { DiskCreatedEvent } from '../events/disk-created.event'
import { DiskStateUpdatedEvent } from '../events/disk-state-updated.event'
import { DiskLastUsedAtUpdatedEvent } from '../events/disk-last-used-at-updated.event'

@EventSubscriber()
export class DiskSubscriber implements EntitySubscriberInterface<Disk> {
  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Disk
  }

  afterInsert(event: InsertEvent<Disk>) {
    this.eventEmitter.emit(DiskEvents.CREATED, new DiskCreatedEvent(event.entity as Disk))
  }

  afterUpdate(event: UpdateEvent<Disk>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    updatedColumns.forEach((column) => {
      switch (column) {
        case 'state':
          this.eventEmitter.emit(
            DiskEvents.STATE_UPDATED,
            new DiskStateUpdatedEvent(event.entity as Disk, event.databaseEntity[column], event.entity[column]),
          )
          break
        case 'lastUsedAt':
          this.eventEmitter.emit(DiskEvents.LAST_USED_AT_UPDATED, new DiskLastUsedAtUpdatedEvent(event.entity as Disk))
          break
        default:
          break
      }
    })
  }
}
