/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { isHealthCheckAuthContext } from '../../common/interfaces/health-check-auth-context.interface'

/**
 * Validates that the current request is authenticated with a `health-check` role auth context.
 */
@Injectable()
export class HealthCheckAuthContextGuard implements CanActivate {
  protected readonly logger = new Logger(HealthCheckAuthContextGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    getAuthContext(context, isHealthCheckAuthContext)
    return true
  }
}
