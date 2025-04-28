/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, UpdateEvent } from 'typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { Workspace } from '../entities/workspace.entity'
import { Inject } from '@nestjs/common'
import { WorkspaceStateUpdatedEvent } from '../events/workspace-state-updated.event'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceDesiredStateUpdatedEvent } from '../events/workspace-desired-state-updated.event'
import { WorkspacePublicStatusUpdatedEvent } from '../events/workspace-public-status-updated.event'
import { WorkspaceCreatedEvent } from '../events/workspace-create.event'
import { WorkspaceOrganizationUpdatedEvent } from '../events/workspace-organization-updated.event'

@EventSubscriber()
export class WorkspaceSubscriber implements EntitySubscriberInterface<Workspace> {
  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return Workspace
  }

  afterInsert(event: InsertEvent<Workspace>) {
    this.eventEmitter.emit(WorkspaceEvents.CREATED, new WorkspaceCreatedEvent(event.entity as Workspace))
  }

  afterUpdate(event: UpdateEvent<Workspace>) {
    const updatedColumns = event.updatedColumns.map((col) => col.propertyName)

    updatedColumns.forEach((column) => {
      switch (column) {
        case 'organizationId':
          this.eventEmitter.emit(
            WorkspaceEvents.ORGANIZATION_UPDATED,
            new WorkspaceOrganizationUpdatedEvent(
              event.entity as Workspace,
              event.databaseEntity[column],
              event.entity[column],
            ),
          )
          break
        case 'public':
          this.eventEmitter.emit(
            WorkspaceEvents.PUBLIC_STATUS_UPDATED,
            new WorkspacePublicStatusUpdatedEvent(
              event.entity as Workspace,
              event.databaseEntity[column],
              event.entity[column],
            ),
          )
          break
        case 'desiredState':
          this.eventEmitter.emit(
            WorkspaceEvents.DESIRED_STATE_UPDATED,
            new WorkspaceDesiredStateUpdatedEvent(
              event.entity as Workspace,
              event.databaseEntity[column],
              event.entity[column],
            ),
          )
          break
        case 'state':
          this.eventEmitter.emit(
            WorkspaceEvents.STATE_UPDATED,
            new WorkspaceStateUpdatedEvent(
              event.entity as Workspace,
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
