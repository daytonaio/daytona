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
import { SandboxService } from '../services/sandbox.service'
import { isBaseAuthContext } from '../../common/interfaces/base-auth-context.interface'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { isRunnerAuthContext } from '../../common/interfaces/runner-auth-context.interface'
import { isRegionAuthContext } from '../../common/interfaces/region-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { isProxyAuthContext } from '../../common/interfaces/proxy-auth-context.interface'
import { isSshGatewayAuthContext } from '../../common/interfaces/ssh-gateway-auth-context.interface'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'

@Injectable()
export class SandboxAccessGuard implements CanActivate {
  private readonly logger = new Logger(SandboxAccessGuard.name)

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
          const sandboxRunnerId = await this.sandboxService.getRunnerId(sandboxIdOrName)
          if (sandboxRunnerId !== authContext.runnerId) {
            throw new ForbiddenException('Runner ID does not match sandbox runner ID')
          }
          break
        }
        case isRegionAuthContext(authContext): {
          const sandboxRegionId = await this.sandboxService.getRegionId(sandboxIdOrName)
          if (sandboxRegionId !== authContext.regionId) {
            throw new ForbiddenException('Sandbox region ID does not match request region ID')
          }
          break
        }
        case isProxyAuthContext(authContext):
        case isSshGatewayAuthContext(authContext):
          break
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
          throw new InvalidAuthenticationContextException()
      }

      // Access granted
      return true
    } catch (error) {
      if (!(error instanceof NotFoundException)) {
        this.logger.error(error)
      }
      throw new NotFoundException(`Sandbox with ID or name ${sandboxIdOrName} not found`)
    }
  }
}
