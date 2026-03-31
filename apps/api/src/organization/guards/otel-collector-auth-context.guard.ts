/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { isOtelCollectorAuthContext } from '../../common/interfaces/otel-collector-auth-context.interface'

/**
 * Validates that the current request is authenticated with an `otel-collector` role auth context.
 */
@Injectable()
export class OtelCollectorAuthContextGuard implements CanActivate {
  protected readonly logger = new Logger(OtelCollectorAuthContextGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    getAuthContext(context, isOtelCollectorAuthContext)
    return true
  }
}
