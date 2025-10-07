/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, Inject, OnModuleInit } from '@nestjs/common'
import { AuditLog } from '../entities/audit-log.entity'
import { AuditLogPublisher } from '../interfaces/audit-publisher.interface'
import { AuditLogStorageAdapter } from '../interfaces/audit-storage.interface'
import { AUDIT_STORAGE_ADAPTER } from '../constants/audit-tokens'

@Injectable()
export class AuditDirectPublisher implements AuditLogPublisher, OnModuleInit {
  private readonly logger = new Logger(AuditDirectPublisher.name)

  constructor(@Inject(AUDIT_STORAGE_ADAPTER) private readonly storageAdapter: AuditLogStorageAdapter) {}

  async onModuleInit(): Promise<void> {
    this.logger.log('Direct storage publisher initialized')
  }

  async write(auditLogs: AuditLog[]): Promise<void> {
    await this.storageAdapter.write(auditLogs)
    this.logger.debug(
      `Written ${auditLogs.length} audit logs directly to ${this.storageAdapter.constructor.name} publisher`,
    )
  }
}
