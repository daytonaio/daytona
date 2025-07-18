/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Inject, Logger } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { DataSource, EntitySubscriberInterface, EventSubscriber, InsertEvent, UpdateEvent } from 'typeorm'
import { AuditLogEvents } from '../constants/audit-log-events.constant'
import { AuditLog } from '../entities/audit-log.entity'
import { AuditLogCreatedEvent } from '../events/audit-log-created.event'
import { AuditLogUpdatedEvent } from '../events/audit-log-updated.event'

@EventSubscriber()
export class AuditLogSubscriber implements EntitySubscriberInterface<AuditLog> {
  private readonly logger = new Logger(AuditLogSubscriber.name)

  @Inject(EventEmitter2)
  private eventEmitter: EventEmitter2

  constructor(dataSource: DataSource) {
    dataSource.subscribers.push(this)
  }

  listenTo() {
    return AuditLog
  }

  beforeUpdate(event: UpdateEvent<AuditLog>) {
    const existingEntity = event.databaseEntity as AuditLog

    if (!existingEntity) {
      // This should not happen, throw exception as a fail-safe
      this.logger.warn('Could not find existing audit log entity, beforeUpdate event:', event)
      throw new ForbiddenException()
    }

    if (existingEntity.statusCode) {
      throw new ForbiddenException('Finalized audit logs are immutable.')
    }
  }

  afterInsert(event: InsertEvent<AuditLog>) {
    this.eventEmitter.emit(AuditLogEvents.CREATED, new AuditLogCreatedEvent(event.entity as AuditLog))
  }

  afterUpdate(event: UpdateEvent<AuditLog>) {
    this.eventEmitter.emit(AuditLogEvents.UPDATED, new AuditLogUpdatedEvent(event.entity as AuditLog))
  }
}
