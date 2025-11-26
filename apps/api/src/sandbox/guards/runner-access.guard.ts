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
import { RegionService } from '../../region/services/region.service'
import { RunnerService } from '../services/runner.service'
import { BaseAuthContext, OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'
import { RegionType } from '../../region/enums/region-type.enum'
import { isRegionProxyContext, RegionProxyContext } from '../../common/interfaces/region-proxy.interface'
import {
  isRegionSSHGatewayContext,
  RegionSSHGatewayContext,
} from '../../common/interfaces/region-ssh-gateway.interface'

@Injectable()
export class RunnerAccessGuard implements CanActivate {
  private readonly logger = new Logger(RunnerAccessGuard.name)

  constructor(
    private readonly runnerService: RunnerService,
    private readonly regionService: RegionService,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const runnerId: string = request.params.runnerId || request.params.id

    // TODO: initialize authContext safely
    const authContext: BaseAuthContext = request.user

    try {
      switch (true) {
        case isRegionProxyContext(authContext): {
          // For region proxy authentication, verify that the runner's region ID matches the proxy's region ID
          const regionProxyContext = authContext as RegionProxyContext
          const runnerRegionId = await this.runnerService.getRegionId(runnerId)
          if (runnerRegionId !== regionProxyContext.regionId) {
            throw new ForbiddenException('Runner region ID does not match region proxy region ID')
          }
          break
        }
        case isRegionSSHGatewayContext(authContext): {
          // For region SSH gateway authentication, verify that the runner's region ID matches the SSH gateway's region ID
          const regionSSHGatewayContext = authContext as RegionSSHGatewayContext
          const runnerRegionId = await this.runnerService.getRegionId(runnerId)
          if (runnerRegionId !== regionSSHGatewayContext.regionId) {
            throw new ForbiddenException('Runner region ID does not match region SSH gateway region ID')
          }
          break
        }
        default: {
          // For user/organization authentication, check organization access
          const orgAuthContext: OrganizationAuthContext = request.user
          const runner = await this.runnerService.findOne(runnerId)
          if (!runner) {
            throw new NotFoundException('Runner not found')
          }

          if (authContext.role !== SystemRole.ADMIN) {
            const region = await this.regionService.findOne(runner.region)
            if (!region) {
              throw new NotFoundException('Region not found')
            }
            if (region.organizationId !== orgAuthContext.organizationId) {
              throw new ForbiddenException('Request organization ID does not match resource organization ID')
            }
            if (region.regionType !== RegionType.CUSTOM) {
              throw new ForbiddenException('Runner is not in a custom region')
            }
          }
        }
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
