/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, ForbiddenException } from '@nestjs/common'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'
import { VolumeService } from '../services/volume.service'
import { Volume } from '../entities/volume.entity'

@Injectable()
export class VolumeAccessGuard implements CanActivate {
  constructor(private readonly volumeService: VolumeService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const volumeIdOrName: string =
      request.params.volumeIdOrName || request.params.name || request.params.volumeId || request.params.id

    const authContext: OrganizationAuthContext = request.user

    let volume: Volume
    try {
      volume = await this.volumeService.findOne(volumeIdOrName)
    } catch (error) {
      volume = await this.volumeService.findByName(authContext.organizationId, volumeIdOrName)
    }

    if (authContext.role !== SystemRole.ADMIN && volume.organizationId !== authContext.organizationId) {
      throw new ForbiddenException('Request organization ID does not match resource organization ID')
    }

    request.volume = volume
    return true
  }
}
