/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { WorkspaceService } from '../services/workspace.service'
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

@Injectable()
export class WorkspaceAccessGuard implements CanActivate {
  constructor(private readonly workspaceService: WorkspaceService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const workspaceId: string = request.params.workspaceId || request.params.id

    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    try {
      const workspace = await this.workspaceService.findOne(workspaceId, true)
      if (authContext.role !== SystemRole.ADMIN && workspace.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      request.workspace = workspace
      return true
    } catch (error) {
      throw new NotFoundException(`Workspace with ID ${workspaceId} not found`)
    }
  }
}
