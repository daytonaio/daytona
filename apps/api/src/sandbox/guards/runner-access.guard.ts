/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, NotFoundException, ForbiddenException, Logger } from '@nestjs/common'
import { ResourceAccessGuard } from '../../common/guards/resource-access.guard'
import { RegionService } from '../../region/services/region.service'
import { RunnerService } from '../services/runner.service'
import { isBaseAuthContext } from '../../common/interfaces/base-auth-context.interface'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { isRegionAuthContext } from '../../common/interfaces/region-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { RegionType } from '../../region/enums/region-type.enum'
import { isProxyAuthContext } from '../../common/interfaces/proxy-auth-context.interface'
import { isSshGatewayAuthContext } from '../../common/interfaces/ssh-gateway-auth-context.interface'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'

@Injectable()
export class RunnerAccessGuard extends ResourceAccessGuard {
  private readonly logger = new Logger(RunnerAccessGuard.name)

  constructor(
    private readonly runnerService: RunnerService,
    private readonly regionService: RegionService,
  ) {
    super()
  }

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const runnerId: string = request.params.runnerId || request.params.id

    const authContext = getAuthContext(context, isBaseAuthContext)

    try {
      const runner = await this.runnerService.findOneOrFail(runnerId)

      switch (true) {
        case isRegionAuthContext(authContext): {
          if (authContext.regionId !== runner.region) {
            throw new ForbiddenException('Region ID does not match runner region ID')
          }
          break
        }
        case isProxyAuthContext(authContext):
        case isSshGatewayAuthContext(authContext):
          break
        case isOrganizationAuthContext(authContext): {
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
          break
        }
        default:
          throw new InvalidAuthenticationContextException()
      }

      // Access granted
      return true
    } catch (error) {
      this.handleResourceAccessError(error, this.logger, 'Runner not found')
    }
  }
}
