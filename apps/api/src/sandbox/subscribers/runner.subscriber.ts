/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, UpdateEvent } from 'typeorm'
import { RunnerEvents } from '../constants/runner-events'
import { Runner } from '../entities/runner.entity'
import { RunnerCreatedEvent } from '../events/runner-created.event'
import { RunnerStateUpdatedEvent } from '../events/runner-state-updated.event'
import { RunnerUnschedulableUpdatedEvent } from '../events/runner-unschedulable-updated.event'

@EventSubscriber()
export class RunnerSubscriber implements EntitySubscriberInterface<Runner> {
  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Runner
  }

  afterInsert(event: InsertEvent<Runner>) {
    this.eventEmitter.emit(RunnerEvents.CREATED, new RunnerCreatedEvent(event.entity as Runner))
  }

  afterUpdate(event: UpdateEvent<Runner>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    updatedColumns.forEach((column) => {
      switch (column) {
        case 'state':
          this.eventEmitter.emit(
            RunnerEvents.STATE_UPDATED,
            new RunnerStateUpdatedEvent(event.entity as Runner, event.databaseEntity[column], event.entity[column]),
          )
          break
        case 'unschedulable':
          this.eventEmitter.emit(
            RunnerEvents.UNSCHEDULABLE_UPDATED,
            new RunnerUnschedulableUpdatedEvent(
              event.entity as Runner,
              event.databaseEntity[column],
              event.entity[column],
            ),
          )
          break
        default:
          break
      }
    })
  }
}
