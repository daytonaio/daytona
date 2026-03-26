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
import { RegionService } from '../services/region.service'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'
import { RegionType } from '../enums/region-type.enum'

@Injectable()
export class RegionAccessGuard implements CanActivate {
  private readonly logger = new Logger(RegionAccessGuard.name)

  constructor(private readonly regionService: RegionService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const regionId: string = request.params.regionId || request.params.id

    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    try {
      const region = await this.regionService.findOne(regionId)
      if (!region) {
        throw new NotFoundException('Region not found')
      }
      if (authContext.role !== SystemRole.ADMIN && region.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      if (authContext.role !== SystemRole.ADMIN && region.regionType !== RegionType.CUSTOM) {
        throw new ForbiddenException('Region is not a custom region')
      }
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException('Region not found')
    }
  }
}
