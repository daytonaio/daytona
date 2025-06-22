/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditLog } from '../entities/audit-log.entity'

export class AuditLogUpdatedEvent {
  constructor(public readonly auditLog: AuditLog) {}
}
