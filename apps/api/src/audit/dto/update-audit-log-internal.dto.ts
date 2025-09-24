/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class UpdateAuditLogInternalDto {
  statusCode?: number | null
  errorMessage?: string | null
  targetId?: string | null
  organizationId?: string | null
}
