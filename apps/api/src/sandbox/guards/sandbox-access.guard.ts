/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { SandboxService } from '../services/sandbox.service'
import { OrganizationAuthContext, BaseAuthContext } from '../../common/interfaces/auth-context.interface'
import { isRunnerContext, RunnerContext } from '../../common/interfaces/runner-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

@Injectable()
export class SandboxAccessGuard implements CanActivate {
  constructor(private readonly sandboxService: SandboxService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    // TODO: remove deprecated request.params.workspaceId param once we remove the deprecated workspace controller
    const sandboxId: string = request.params.sandboxId || request.params.id || request.params.workspaceId

    // TODO: initialize authContext safely
    const authContext: BaseAuthContext = request.user

    try {
      const sandbox = await this.sandboxService.findOne(sandboxId, true)

      // Check if this is a runner making the request
      if (isRunnerContext(authContext)) {
        // For runner authentication, verify that the runner ID matches the sandbox's runner ID
        const runnerContext = authContext as RunnerContext
        if (sandbox.runnerId !== runnerContext.runnerId) {
          throw new ForbiddenException('Runner ID does not match sandbox runner ID')
        }
      } else {
        // For user/organization authentication, check organization access
        const orgAuthContext = authContext as OrganizationAuthContext
        if (
          orgAuthContext.role !== 'ssh-gateway' &&
          orgAuthContext.role !== SystemRole.ADMIN &&
          sandbox.organizationId !== orgAuthContext.organizationId
        ) {
          throw new ForbiddenException('Request organization ID does not match resource organization ID')
        }
      }

      request.sandbox = sandbox
      return true
    } catch (error) {
      if (error instanceof ForbiddenException) {
        throw error
      }
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }
  }
}
