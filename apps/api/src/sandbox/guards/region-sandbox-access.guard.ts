/*
 * Copyright Daytona Platforms Inc.
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
import { SandboxService } from '../services/sandbox.service'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { isRegionAuthContext } from '../../common/interfaces/region-auth-context.interface'

@Injectable()
export class RegionSandboxAccessGuard implements CanActivate {
  private readonly logger = new Logger(RegionSandboxAccessGuard.name)

  constructor(private readonly sandboxService: SandboxService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const sandboxId: string = request.params.sandboxId || request.params.id
    const authContext = getAuthContext(context, isRegionAuthContext)

    try {
      const sandboxRegionId = await this.sandboxService.getRegionId(sandboxId)
      if (sandboxRegionId !== authContext.regionId) {
        throw new ForbiddenException(`Sandbox region ID does not match region ${authContext.role} region ID`)
      }
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException(`Sandbox with ID or name ${sandboxId} not found`)
    }
  }
}
