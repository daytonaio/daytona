/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, Logger } from '@nestjs/common'
import { isDaemonContext } from '../common/interfaces/daemon-context.interface'
import { getAuthContext } from './get-auth-context'

@Injectable()
export class DaemonAuthGuard implements CanActivate {
  private readonly logger = new Logger(DaemonAuthGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    // Throws if not daemon context
    getAuthContext(context, isDaemonContext)
    return true
  }
}
