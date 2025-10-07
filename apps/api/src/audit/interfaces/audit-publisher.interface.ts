/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditLog } from '../entities/audit-log.entity'

/**
 * Interface for audit log publisher operations
 * Handles publishing audit logs
 */
export interface AuditLogPublisher {
  /**
   * Publish audit logs
   */
  write(auditLogs: AuditLog[]): Promise<void>
}
