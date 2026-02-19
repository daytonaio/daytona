/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { getAuthContext } from './get-auth-context'
import { isHealthCheckContext } from '../common/interfaces/health-check-context.interface'

@Injectable()
export class HealthCheckGuard implements CanActivate {
  protected readonly logger = new Logger(HealthCheckGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    // Throws if not health check context
    getAuthContext(context, isHealthCheckContext)
    return true
  }
}
