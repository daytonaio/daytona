/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { getAuthContext } from './get-auth-context'
import { isOtelProxyContext } from '../common/interfaces/otel-proxy-context.interface'

@Injectable()
export class OtelProxyGuard implements CanActivate {
  protected readonly logger = new Logger(OtelProxyGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    // Throws if not proxy context
    getAuthContext(context, isOtelProxyContext)
    return true
  }
}
