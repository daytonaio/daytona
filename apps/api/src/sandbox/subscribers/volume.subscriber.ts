/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, UpdateEvent } from 'typeorm'
import { VolumeEvents } from '../constants/volume-events'
import { Volume } from '../entities/volume.entity'
import { VolumeCreatedEvent } from '../events/volume-created.event'
import { VolumeStateUpdatedEvent } from '../events/volume-state-updated.event'
import { VolumeLastUsedAtUpdatedEvent } from '../events/volume-last-used-at-updated.event'

@EventSubscriber()
export class VolumeSubscriber implements EntitySubscriberInterface<Volume> {
  constructor(
    dataSource: DataSource,
    @Inject(EventEmitter2)
    private eventEmitter: EventEmitter2,
  ) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Volume
  }

  afterInsert(event: InsertEvent<Volume>) {
    this.eventEmitter.emit(VolumeEvents.CREATED, new VolumeCreatedEvent(event.entity as Volume))
  }

  afterUpdate(event: UpdateEvent<Volume>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    if (!event.entity) {
      return
    }

    const entity = event.entity as Volume

    updatedColumns.forEach((column) => {
      switch (column) {
        case 'state':
          this.eventEmitter.emit(
            VolumeEvents.STATE_UPDATED,
            new VolumeStateUpdatedEvent(entity, event.databaseEntity[column], entity[column]),
          )
          break
        case 'lastUsedAt':
          this.eventEmitter.emit(
            VolumeEvents.LAST_USED_AT_UPDATED,
            new VolumeLastUsedAtUpdatedEvent(event.entity as Volume),
          )
          break
        default:
          break
      }
    })
  }
}
