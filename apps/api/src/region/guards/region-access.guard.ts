/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, NotFoundException, ForbiddenException, Logger } from '@nestjs/common'
import { ResourceAccessGuard } from '../../common/guards/resource-access.guard'
import { RegionService } from '../services/region.service'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { RegionType } from '../enums/region-type.enum'
import { EntityNotFoundError } from 'typeorm'

@Injectable()
export class RegionAccessGuard extends ResourceAccessGuard {
  private readonly logger = new Logger(RegionAccessGuard.name)

  constructor(private readonly regionService: RegionService) {
    super()
  }

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const regionId: string = request.params.regionId || request.params.id

    const authContext = getAuthContext(context, isOrganizationAuthContext)

    try {
      const region = await this.regionService.findOne(regionId)
      if (!region) {
        throw new NotFoundException('Region not found')
      }
      if (region.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      if (region.regionType !== RegionType.CUSTOM) {
        throw new ForbiddenException('Region is not a custom region')
      }
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException) && !(error instanceof EntityNotFoundError)) {
        this.logger.error(error)
      }
      throw new NotFoundException('Region not found')
    }
  }
}
