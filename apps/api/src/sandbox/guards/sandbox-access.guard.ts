/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { SandboxService } from '../services/sandbox.service'
import { SystemRole } from '../../user/enums/system-role.enum'
import { RequestWithOrganizationContext } from '../../common/types/request.types'

@Injectable()
export class SandboxAccessGuard implements CanActivate {
  constructor(private readonly sandboxService: SandboxService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest<RequestWithOrganizationContext>()
    // TODO: remove deprecated request.params.workspaceId param once we remove the deprecated workspace controller
    const sandboxId: string = request.params.sandboxId || request.params.id || request.params.workspaceId

    const authContext = request.user

    try {
      const sandbox = await this.sandboxService.findOne(sandboxId, true)
      if (authContext.role !== SystemRole.ADMIN && sandbox.organizationId !== authContext.organizationId) {
        throw new ForbiddenException('Request organization ID does not match resource organization ID')
      }
      request.sandbox = sandbox
      return true
    } catch (error) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }
  }
}
