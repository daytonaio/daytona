/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { getAuthContext } from './get-auth-context'
import { isOtelCollectorContext } from '../common/interfaces/otel-collector-context.interface'

@Injectable()
export class OtelCollectorGuard implements CanActivate {
  protected readonly logger = new Logger(OtelCollectorGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    // Throws if not proxy context
    getAuthContext(context, isOtelCollectorContext)
    return true
  }
}
