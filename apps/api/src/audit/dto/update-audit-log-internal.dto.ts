/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuditOutcome } from '../enums/audit-outcome-enum'

export class UpdateAuditLogInternalDto {
  outcome: AuditOutcome
  errorMessage?: string
  targetId?: string
}
