/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CanActivate, ExecutionContext, Injectable, Logger } from '@nestjs/common'
import { OrganizationService } from '../services/organization.service'
import { OrganizationUserService } from '../services/organization-user.service'
import { AuthContext, OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

@Injectable()
export class OrganizationAccessGuard implements CanActivate {
  protected readonly logger = new Logger(OrganizationAccessGuard.name)

  constructor(
    private readonly organizationService: OrganizationService,
    private readonly organizationUserService: OrganizationUserService,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    // TODO: initialize authContext safely
    const authContext: AuthContext = request.user

    if (!authContext) {
      this.logger.warn('User object is undefined. Authentication may not be set up correctly.')
      return false
    }

    // note: semantic parameter names must be used (avoid :id)
    const organizationIdParam = request.params.organizationId || request.params.orgId

    if (!organizationIdParam && !authContext.organizationId) {
      this.logger.warn('Organization ID missing from the request context.')
      return false
    }

    if (
      organizationIdParam &&
      authContext.apiKey &&
      authContext.apiKey.organizationId !== organizationIdParam &&
      authContext.role !== SystemRole.ADMIN
    ) {
      this.logger.warn(
        `Organization ID mismatch in the request context. Expected: ${organizationIdParam}, Actual: ${authContext.apiKey.organizationId}`,
      )
      this.logger.warn('Organization ID mismatch in the request context.')
      return false
    }

    const organizationId = organizationIdParam || authContext.organizationId

    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      this.logger.warn(`Organization not found. Organization ID: ${organizationId}`)
      return false
    }

    const organizationAuthContext: OrganizationAuthContext = {
      ...authContext,
      organizationId,
      organization,
    }
    request.user = organizationAuthContext

    if (authContext.role === SystemRole.ADMIN) {
      return true
    }

    const organizationUser = await this.organizationUserService.findOne(organizationId, authContext.userId)
    if (!organizationUser) {
      this.logger.warn(
        `Organization user not found. User ID: ${authContext.userId}, Organization ID: ${organizationId}`,
      )
      return false
    }

    organizationAuthContext.organizationUser = organizationUser
    request.user = organizationAuthContext

    return true
  }
}
