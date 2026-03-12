/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { OrganizationAccessGuard } from './organization-access.guard'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationService } from '../services/organization.service'
import { OrganizationUserService } from '../services/organization-user.service'
import { isOrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'

@Injectable()
export class OrganizationActionGuard extends OrganizationAccessGuard {
  protected readonly logger = new Logger(OrganizationActionGuard.name)

  constructor(
    organizationService: OrganizationService,
    organizationUserService: OrganizationUserService,
    private readonly reflector: Reflector,
  ) {
    super(organizationService, organizationUserService)
  }

  async canActivate(context: ExecutionContext): Promise<boolean> {
    if (!(await super.canActivate(context))) {
      return false
    }

    const authContext = getAuthContext(context, isOrganizationAuthContext)

    if (!authContext.organizationUser) {
      return false
    }

    const requiredRole = this.reflector.get(RequiredOrganizationMemberRole, context.getHandler())
    if (!requiredRole) {
      return true
    }

    if (requiredRole === OrganizationMemberRole.OWNER) {
      return authContext.organizationUser.role === OrganizationMemberRole.OWNER
    }

    return true
  }
}
