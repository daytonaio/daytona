/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource, EntitySubscriberInterface, EventSubscriber, UpdateEvent } from 'typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { Sandbox } from '../entities/sandbox.entity'
import { Inject, Logger } from '@nestjs/common'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxDesiredStateUpdatedEvent } from '../events/sandbox-desired-state-updated.event'
import { SandboxPublicStatusUpdatedEvent } from '../events/sandbox-public-status-updated.event'
import { SandboxOrganizationUpdatedEvent } from '../events/sandbox-organization-updated.event'
import { SandboxLookupCacheInvalidationService } from '../services/sandbox-lookup-cache-invalidation.service'

@EventSubscriber()
export class SandboxSubscriber implements EntitySubscriberInterface<Sandbox> {
  private readonly logger = new Logger(SandboxSubscriber.name)

  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  @Inject(SandboxLookupCacheInvalidationService)
  private sandboxLookupCacheInvalidationService: SandboxLookupCacheInvalidationService

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Sandbox
  }

  afterUpdate(event: UpdateEvent<Sandbox>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    updatedColumns.forEach((column) => {
      // For QueryBuilder/Repository.update(), TypeORM doesn't provide databaseEntity.
      if (!event.entity || !event.databaseEntity) {
        return
      }

      switch (column) {
        case 'organizationId':
          this.eventEmitter.emit(
            SandboxEvents.ORGANIZATION_UPDATED,
            new SandboxOrganizationUpdatedEvent(
              event.entity as Sandbox,
              event.databaseEntity[column],
              event.entity[column],
            ),
          )
          break
        case 'public':
          this.eventEmitter.emit(
            SandboxEvents.PUBLIC_STATUS_UPDATED,
            new SandboxPublicStatusUpdatedEvent(
              event.entity as Sandbox,
              event.databaseEntity[column],
              event.entity[column],
            ),
          )
          break
        case 'desiredState':
          this.eventEmitter.emit(
            SandboxEvents.DESIRED_STATE_UPDATED,
            new SandboxDesiredStateUpdatedEvent(
              event.entity as Sandbox,
              event.databaseEntity[column],
              event.entity[column],
            ),
          )
          break
        case 'state':
          this.eventEmitter.emit(
            SandboxEvents.STATE_UPDATED,
            new SandboxStateUpdatedEvent(event.entity as Sandbox, event.databaseEntity[column], event.entity[column]),
          )
          break
        default:
          break
      }
    })

    // Invalidate cached sandbox lookup queries (by id / by name) on any update triggered via save().
    // Note: Repository.update() does not provide databaseEntity, so those paths should invalidate explicitly.
    const entity = event.entity as Sandbox | undefined
    const dbEntity = event.databaseEntity as Sandbox | undefined
    if (!entity || !dbEntity) {
      return
    }

    try {
      this.sandboxLookupCacheInvalidationService.invalidate({
        sandboxId: entity.id,
        organizationId: entity.organizationId,
        name: entity.name,
        previousOrganizationId: dbEntity.organizationId,
        previousName: dbEntity.name,
      })
    } catch (error) {
      this.logger.warn(
        `Failed to enqueue sandbox lookup cache invalidation for ${entity.id}: ${error instanceof Error ? error.message : String(error)}`,
      )
    }
  }
}
