/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger } from '@nestjs/common'
import { AuthContextGuard } from '../../common/guards/auth-context.guard'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { isUserAuthContext } from '../../common/interfaces/user-auth-context.interface'

/**
 * Validates that the current request is authenticated with a user auth context.
 */
@Injectable()
export class UserAuthContextGuard extends AuthContextGuard {
  private readonly logger = new Logger(UserAuthContextGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    getAuthContext(context, isUserAuthContext)
    return true
  }
}
