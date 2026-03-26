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
import { OrganizationAuthContext } from '../../common/interfaces/auth-context.interface'
import { SystemRole } from '../../user/enums/system-role.enum'

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

    const request = context.switchToHttp().getRequest()
    // TODO: initialize authContext safely
    const authContext: OrganizationAuthContext = request.user

    if (authContext.role === SystemRole.ADMIN) {
      return true
    }

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
