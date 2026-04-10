/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger } from '@nestjs/common'
import { AuthContextGuard } from '../../common/guards/auth-context.guard'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { ProxyAuthContext, isProxyAuthContext } from '../../common/interfaces/proxy-auth-context.interface'
import {
  RegionProxyAuthContext,
  isRegionProxyAuthContext,
} from '../../common/interfaces/region-proxy-auth-context.interface'

/**
 * Validates that the current request is authenticated with a `proxy` or `region-proxy` role auth context.
 */
@Injectable()
export class ProxyAuthContextGuard extends AuthContextGuard {
  protected readonly logger = new Logger(ProxyAuthContextGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    getAuthContext(
      context,
      (user: unknown): user is ProxyAuthContext | RegionProxyAuthContext =>
        isProxyAuthContext(user) || isRegionProxyAuthContext(user),
    )
    return true
  }
}
