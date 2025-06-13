/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { getAuthContext } from './get-auth-context'
import { isProxyContext } from '../common/interfaces/proxy-context.interface'

@Injectable()
export class ProxyGuard implements CanActivate {
  protected readonly logger = new Logger(ProxyGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    // Throws if not proxy context
    getAuthContext(context, isProxyContext)
    return true
  }
}
