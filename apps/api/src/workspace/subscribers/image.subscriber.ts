/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, RemoveEvent, UpdateEvent } from 'typeorm'
import { ImageEvents } from '../constants/image-events'
import { Image } from '../entities/image.entity'
import { ImageCreatedEvent } from '../events/image-created.event'
import { ImageStateUpdatedEvent } from '../events/image-state-updated.event'
import { ImageEnabledToggledEvent } from '../events/image-enabled-toggled.event'
import { ImageRemovedEvent } from '../events/image-removed.event'

@EventSubscriber()
export class ImageSubscriber implements EntitySubscriberInterface<Image> {
  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Image
  }

  afterInsert(event: InsertEvent<Image>) {
    this.eventEmitter.emit(ImageEvents.CREATED, new ImageCreatedEvent(event.entity as Image))
  }

  afterUpdate(event: UpdateEvent<Image>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    updatedColumns.forEach((column) => {
      switch (column) {
        case 'enabled':
          this.eventEmitter.emit(ImageEvents.ENABLED_TOGGLED, new ImageEnabledToggledEvent(event.entity as Image))
          break
        case 'state':
          this.eventEmitter.emit(
            ImageEvents.STATE_UPDATED,
            new ImageStateUpdatedEvent(event.entity as Image, event.databaseEntity[column], event.entity[column]),
          )
          break
        default:
          break
      }
    })
  }

  afterRemove(event: RemoveEvent<Image>) {
    this.eventEmitter.emit(ImageEvents.REMOVED, new ImageRemovedEvent(event.databaseEntity as Image))
  }
}
