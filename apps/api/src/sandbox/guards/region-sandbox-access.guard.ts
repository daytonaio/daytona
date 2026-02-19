/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { SandboxService } from '../services/sandbox.service'
import { BaseAuthContext } from '../../common/interfaces/auth-context.interface'
import { isRegionProxyContext, RegionProxyContext } from '../../common/interfaces/region-proxy.interface'
import {
  isRegionSSHGatewayContext,
  RegionSSHGatewayContext,
} from '../../common/interfaces/region-ssh-gateway.interface'

@Injectable()
export class RegionSandboxAccessGuard implements CanActivate {
  constructor(private readonly sandboxService: SandboxService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const sandboxId: string = request.params.sandboxId || request.params.id

    const authContext: BaseAuthContext = request.user

    if (!isRegionProxyContext(authContext) && !isRegionSSHGatewayContext(authContext)) {
      return false
    }

    try {
      const regionContext = authContext as RegionProxyContext | RegionSSHGatewayContext
      const sandboxRegionId = await this.sandboxService.getRegionId(sandboxId)
      if (sandboxRegionId !== regionContext.regionId) {
        throw new ForbiddenException(`Sandbox region ID does not match region ${regionContext.role} region ID`)
      }
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        console.error(error)
      }
      throw new NotFoundException(`Sandbox with ID or name ${sandboxId} not found`)
    }
  }
}
