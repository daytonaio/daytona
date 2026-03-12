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
import { getAuthContext } from '../../common/utils/get-auth-context'
import { isRegionAuthContext } from '../../common/interfaces/region-auth-context.interface'

@Injectable()
export class RegionRunnerAccessGuard implements CanActivate {
  private readonly logger = new Logger(RegionRunnerAccessGuard.name)

  constructor(private readonly runnerService: RunnerService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const runnerId: string = request.params.runnerId || request.params.id
    const authContext = getAuthContext(context, isRegionAuthContext)

    try {
      const runner = await this.runnerService.findOneOrFail(runnerId)
      if (authContext.regionId !== runner.region) {
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
