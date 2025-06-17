/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SetMetadata } from '@nestjs/common'
import { AuditAction } from '../enums/audit-action.enum'
import { AuditTarget } from '../enums/audit-target.enum'

// TODO: body param resolver
export interface AuditMetadata {
  action: AuditAction
  targetType: AuditTarget
  targetIdParam?: string
  targetIdResolver?: (result: any) => string | null | undefined
}

export const AUDIT_METADATA_KEY = 'audit_metadata'

export const Audit = (metadata: AuditMetadata) => SetMetadata(AUDIT_METADATA_KEY, metadata)
