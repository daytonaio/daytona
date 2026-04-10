/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger } from '@nestjs/common'
import { AuthContextGuard } from '../../common/guards/auth-context.guard'
import { getAuthContext } from '../../common/utils/get-auth-context'
import {
  SshGatewayAuthContext,
  isSshGatewayAuthContext,
} from '../../common/interfaces/ssh-gateway-auth-context.interface'
import {
  RegionSSHGatewayAuthContext,
  isRegionSSHGatewayAuthContext,
} from '../../common/interfaces/region-ssh-gateway-auth-context.interface'

/**
 * Validates that the current request is authenticated with a `ssh-gateway` or `region-ssh-gateway` role auth context.
 */
@Injectable()
export class SshGatewayAuthContextGuard extends AuthContextGuard {
  protected readonly logger = new Logger(SshGatewayAuthContextGuard.name)

  async canActivate(context: ExecutionContext): Promise<boolean> {
    getAuthContext(
      context,
      (user: unknown): user is SshGatewayAuthContext | RegionSSHGatewayAuthContext =>
        isSshGatewayAuthContext(user) || isRegionSSHGatewayAuthContext(user),
    )
    return true
  }
}
