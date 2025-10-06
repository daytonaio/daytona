/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Logger } from '@nestjs/common'
import { DataSource, EntitySubscriberInterface, EventSubscriber, UpdateEvent } from 'typeorm'
import { AuditLog } from '../entities/audit-log.entity'

@EventSubscriber()
export class AuditLogSubscriber implements EntitySubscriberInterface<AuditLog> {
  private readonly logger = new Logger(AuditLogSubscriber.name)

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
}
