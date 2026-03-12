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
import { isBaseAuthContext } from '../../common/interfaces/auth-context.interface'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { SystemRole } from '../../user/enums/system-role.enum'
import { RegionType } from '../../region/enums/region-type.enum'
import { isRegionProxyAuthContext } from '../../common/interfaces/region-proxy-auth-context.interface'
import { isRegionSSHGatewayAuthContext } from '../../common/interfaces/region-ssh-gateway-auth-context.interface'
import { isProxyAuthContext } from '../../common/interfaces/proxy-auth-context.interface'
import { isSshGatewayAuthContext } from '../../common/interfaces/ssh-gateway-auth-context.interface'

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

    const authContext = getAuthContext(context, isBaseAuthContext)

    try {
      const runner = await this.runnerService.findOneOrFail(runnerId)

      switch (true) {
        case isRegionProxyAuthContext(authContext):
        case isRegionSSHGatewayAuthContext(authContext): {
          // Use RunnerRegionAccessGuard to check access instead
          return false
        }
        case isProxyAuthContext(authContext):
        case isSshGatewayAuthContext(authContext):
          return true
        case isOrganizationAuthContext(authContext): {
          if (authContext.role !== SystemRole.ADMIN) {
            const region = await this.regionService.findOne(runner.region)
            if (!region) {
              throw new NotFoundException('Region not found')
            }
            if (region.organizationId !== authContext.organizationId) {
              throw new ForbiddenException('Request organization ID does not match resource organization ID')
            }
            if (region.regionType !== RegionType.CUSTOM) {
              throw new ForbiddenException('Runner is not in a custom region')
            }
          }
          return true
        }
        default:
          return false
      }
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException('Runner not found')
    }
  }
}
