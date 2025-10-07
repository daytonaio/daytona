/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditLog } from '../entities/audit-log.entity'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import { AuditLogFilter } from './audit-filter.interface'

/**
 * Interface for audit log storage operations
 * Handles persistent storage and audit logs queries
 */
export interface AuditLogStorageAdapter {
  /**
   * Write audit logs to storage
   */
  write(auditLogs: AuditLog[]): Promise<void>

  /**
   * Get all audit logs
   */
  getAllLogs(page?: number, limit?: number, filters?: AuditLogFilter): Promise<PaginatedList<AuditLog>>

  /**
   * Get audit logs for organization
   */
  getOrganizationLogs(
    organizationId: string,
    page?: number,
    limit?: number,
    filters?: AuditLogFilter,
  ): Promise<PaginatedList<AuditLog>>
}
