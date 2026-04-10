/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger } from '@nestjs/common'
import { AuthContextGuard } from '../../common/guards/auth-context.guard'
import { isRunnerAuthContext } from '../../common/interfaces/runner-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'

/**
 * Validates that the current request is authenticated with a `runner` role auth context.
 */
@Injectable()
export class RunnerAuthContextGuard extends AuthContextGuard {
  private readonly logger = new Logger(RunnerAuthContextGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    getAuthContext(context, isRunnerAuthContext)
    return true
  }
}
