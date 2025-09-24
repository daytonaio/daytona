/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext } from '@nestjs/common'
import { isRunnerContext } from '../common/interfaces/runner-context.interface'
import { getAuthContext } from './get-auth-context'

@Injectable()
export class RunnerAuthGuard implements CanActivate {
  async canActivate(context: ExecutionContext): Promise<boolean> {
    // Throws if not runner context
    getAuthContext(context, isRunnerContext)
    return true
  }
}
