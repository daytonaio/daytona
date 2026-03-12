/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, CanActivate, ExecutionContext, NotFoundException, ForbiddenException } from '@nestjs/common'
import { SandboxService } from '../services/sandbox.service'
import { isBaseAuthContext } from '../../common/interfaces/auth-context.interface'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { isRunnerAuthContext } from '../../common/interfaces/runner-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { isProxyAuthContext } from '../../common/interfaces/proxy-auth-context.interface'
import { isSshGatewayAuthContext } from '../../common/interfaces/ssh-gateway-auth-context.interface'
import { isRegionProxyAuthContext } from '../../common/interfaces/region-proxy-auth-context.interface'
import { isRegionSSHGatewayAuthContext } from '../../common/interfaces/region-ssh-gateway-auth-context.interface'

@Injectable()
export class SandboxAccessGuard implements CanActivate {
  constructor(private readonly sandboxService: SandboxService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    // TODO: remove deprecated request.params.workspaceId param once we remove the deprecated workspace controller
    const sandboxIdOrName: string =
      request.params.sandboxIdOrName || request.params.sandboxId || request.params.id || request.params.workspaceId

    const authContext = getAuthContext(context, isBaseAuthContext)

    try {
      switch (true) {
        case isRunnerAuthContext(authContext): {
          // For runner authentication, verify that the runner ID matches the sandbox's runner ID
          const sandboxRunnerId = await this.sandboxService.getRunnerId(sandboxIdOrName)
          if (sandboxRunnerId !== authContext.runnerId) {
            throw new ForbiddenException('Runner ID does not match sandbox runner ID')
          }
          break
        }
        case isRegionProxyAuthContext(authContext):
        case isRegionSSHGatewayAuthContext(authContext): {
          // Use RegionSandboxAccessGuard to check access instead
          return false
        }
        case isProxyAuthContext(authContext):
        case isSshGatewayAuthContext(authContext):
          return true
        case isOrganizationAuthContext(authContext): {
          const sandboxOrganizationId = await this.sandboxService.getOrganizationId(
            sandboxIdOrName,
            authContext.organizationId,
          )
          if (sandboxOrganizationId !== authContext.organizationId) {
            throw new ForbiddenException('Request organization ID does not match resource organization ID')
          }
          break
        }
        default:
          return false
      }
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        console.error(error)
      }
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} not found`)
    }
  }
}
