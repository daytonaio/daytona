/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Injectable,
  CanActivate,
  ExecutionContext,
  NotFoundException,
  ForbiddenException,
  Logger,
} from '@nestjs/common'
import { RunnerService } from '../services/runner.service'
import { BaseAuthContext } from '../../common/interfaces/auth-context.interface'
import { isRegionProxyContext, RegionProxyContext } from '../../common/interfaces/region-proxy.interface'
import {
  isRegionSSHGatewayContext,
  RegionSSHGatewayContext,
} from '../../common/interfaces/region-ssh-gateway.interface'

@Injectable()
export class RegionRunnerAccessGuard implements CanActivate {
  private readonly logger = new Logger(RegionRunnerAccessGuard.name)

  constructor(private readonly runnerService: RunnerService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const runnerId: string = request.params.runnerId || request.params.id

    const authContext: BaseAuthContext = request.user

    if (!isRegionProxyContext(authContext) && !isRegionSSHGatewayContext(authContext)) {
      return false
    }

    try {
      const regionContext = authContext as RegionProxyContext | RegionSSHGatewayContext
      const runner = await this.runnerService.findOneOrFail(runnerId)
      if (regionContext.regionId !== runner.region) {
        throw new ForbiddenException('Region ID does not match runner region ID')
      }
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException('Runner not found')
    }
  }
}
