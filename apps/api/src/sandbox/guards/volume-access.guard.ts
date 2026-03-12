/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { Injectable, CanActivate, ExecutionContext, ForbiddenException, NotFoundException } from '@nestjs/common'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { SystemRole } from '../../user/enums/system-role.enum'
import { VolumeService } from '../services/volume.service'

@Injectable()
export class VolumeAccessGuard implements CanActivate {
  constructor(private readonly volumeService: VolumeService) {}
  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()

    const volumeId = request.params.volumeId || request.params.id
    const volumeName = request.params.name

    if (!volumeId && !volumeName) {
      throw new NotFoundException(`Volume not found`)
    }

    const authContext = getAuthContext(context, isOrganizationAuthContext)

    try {
      const params = volumeId ? { id: volumeId } : { name: volumeName, organizationId: authContext.organizationId }
      const volumeOrganizationId = await this.volumeService.getOrganizationId(params)

      if (authContext.role !== SystemRole.ADMIN && volumeOrganizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
    } catch {
      throw new NotFoundException(`Volume with ${volumeId ? 'ID' : 'name'} ${volumeId || volumeName} not found`)
    }

    return true
  }
}
