/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, UpdateEvent } from 'typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { Sandbox } from '../entities/sandbox.entity'
import { Inject } from '@nestjs/common'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxDesiredStateUpdatedEvent } from '../events/sandbox-desired-state-updated.event'
import { SandboxPublicStatusUpdatedEvent } from '../events/sandbox-public-status-updated.event'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { SandboxOrganizationUpdatedEvent } from '../events/sandbox-organization-updated.event'

@EventSubscriber()
export class SandboxSubscriber implements EntitySubscriberInterface<Sandbox> {
  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Sandbox
  }

  afterInsert(event: InsertEvent<Sandbox>) {
    this.eventEmitter.emit(SandboxEvents.CREATED, new SandboxCreatedEvent(event.entity as Sandbox))
  }

  afterUpdate(event: UpdateEvent<Sandbox>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    updatedColumns.forEach((column) => {
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
  }
}
