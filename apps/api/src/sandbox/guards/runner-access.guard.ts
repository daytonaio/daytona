/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { RegionService } from '../../region/services/region.service'
import { RunnerService } from '../services/runner.service'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

@Injectable()
export class RunnerAccessGuard implements CanActivate {
  constructor(
    private readonly runnerService: RunnerService,
    private readonly regionService: RegionService,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const runnerId: string = request.params.runnerId || request.params.id

    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    try {
      const runnerRegionId = await this.runnerService.getRegionId(runnerId)
      if (authContext.role !== SystemRole.ADMIN) {
        const regionOrganizationId = await this.regionService.getOrganizationId(runnerRegionId)
        if (regionOrganizationId !== authContext.organizationId) {
          throw new ForbiddenException('Request organization ID does not match resource organization ID')
        }
      }
      return true
    } catch (error) {
      throw new NotFoundException(`Runner with ID ${runnerId} not found`)
    }
  }
}
