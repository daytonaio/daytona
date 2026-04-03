/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExecutionContext, Injectable } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { CombinedAuthGuard } from './combined-auth.guard'
import { IS_PUBLIC_KEY } from './decorators/public.decorator'

@Injectable()
export class GlobalAuthGuard extends CombinedAuthGuard {
  constructor(private readonly reflector: Reflector) {
    super()
  }

  canActivate(context: ExecutionContext) {
    // Only enforce this guard for HTTP routes.
    if (context.getType<'http' | 'rpc' | 'ws'>() !== 'http') {
      return true
    }

    const isPublic = this.reflector.getAllAndOverride<boolean>(IS_PUBLIC_KEY, [
      context.getHandler(),
      context.getClass(),
    ])

    if (isPublic) {
      return true
    }

    return super.canActivate(context)
  }
}
