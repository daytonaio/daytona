/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SetMetadata } from '@nestjs/common'
import { Request } from 'express'
import { AuditAction } from '../enums/audit-action.enum'
import { AuditTarget } from '../enums/audit-target.enum'

export type TypedRequest<T> = Omit<Request, 'body'> & { body: T }

export interface AuditContext {
  action: AuditAction
  targetType?: AuditTarget
  // TODO: also a resolver function?
  targetIdParam?: string
  targetIdResolver?: (result: any) => string | null | undefined
  metadata?: Record<string, (req: Request) => any> | Record<string, any>
}

export const AUDIT_CONTEXT_KEY = 'audit_context'

export const Audit = (context: AuditContext) => SetMetadata(AUDIT_CONTEXT_KEY, context)
