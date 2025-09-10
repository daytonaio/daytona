/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { RegionService } from '../services/region.service'
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
      const runner = await this.runnerService.findOne(runnerId)
      if (authContext.role !== SystemRole.ADMIN) {
        const region = await this.regionService.findOne(runner.regionId)
        if (region.organizationId !== authContext.organizationId) {
          throw new ForbiddenException('Request organization ID does not match resource organization ID')
        }
      }
      request.runner = runner
      return true
    } catch (error) {
      throw new NotFoundException(`Runner with ID ${runnerId} not found`)
    }
  }
}
