/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, ForbiddenException, NotFoundException } from '@nestjs/common'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'
import { VolumeService } from '../services/volume.service'

@Injectable()
export class VolumeAccessGuard implements CanActivate {
  constructor(private readonly volumeService: VolumeService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const volumeIdOrName: string =
      request.params.volumeIdOrName || request.params.name || request.params.volumeId || request.params.id

    const authContext: OrganizationAuthContext = request.user

    try {
      const volumeOrganizationId = await this.volumeService.getOrganizationId(
        volumeIdOrName,
        authContext.organizationId,
      )

      if (authContext.role !== SystemRole.ADMIN && volumeOrganizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
    } catch {
      throw new NotFoundException(`Volume with ID or name ${volumeIdOrName} not found`)
    }

    return true
  }
}
