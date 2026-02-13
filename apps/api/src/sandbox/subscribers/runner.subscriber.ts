/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject, Logger } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, UpdateEvent } from 'typeorm'
import { RunnerEvents } from '../constants/runner-events'
import { Runner } from '../entities/runner.entity'
import { RunnerCreatedEvent } from '../events/runner-created.event'
import { RunnerStateUpdatedEvent } from '../events/runner-state-updated.event'
import { RunnerUnschedulableUpdatedEvent } from '../events/runner-unschedulable-updated.event'
import { runnerLookupCacheKeyById } from '../utils/runner-lookup-cache.util'

@EventSubscriber()
export class RunnerSubscriber implements EntitySubscriberInterface<Runner> {
  private readonly logger = new Logger(RunnerSubscriber.name)

  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  private dataSource: DataSource

  constructor(dataSource: DataSource) {
    this.dataSource = dataSource
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
      // For Repository.update(), TypeORM doesn't provide databaseEntity.
      if (!event.entity || !event.databaseEntity) {
        return
      }

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

    // Invalidate cached runner lookup queries on any update triggered via save().
    // Note: Repository.update() does not provide databaseEntity, so those paths
    // invalidate explicitly via RunnerService.updateRunner().
    const entity = event.entity as Runner | undefined
    if (!entity?.id) {
      return
    }

    const cache = this.dataSource.queryResultCache
    if (!cache) {
      return
    }

    cache
      .remove([runnerLookupCacheKeyById(entity.id)])
      .catch((error) =>
        this.logger.warn(
          `Failed to invalidate runner lookup cache for ${entity.id}: ${error instanceof Error ? error.message : String(error)}`,
        ),
      )
  }
}
