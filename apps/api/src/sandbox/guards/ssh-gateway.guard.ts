/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { getAuthContext } from '../../auth/get-auth-context'
import { isSshGatewayContext } from '../../common/interfaces/ssh-gateway-context.interface'

@Injectable()
export class SshGatewayGuard implements CanActivate {
  protected readonly logger = new Logger(SshGatewayGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    // Throws if not ssh gateway context
    getAuthContext(context, isSshGatewayContext)
    return true
  }
}
