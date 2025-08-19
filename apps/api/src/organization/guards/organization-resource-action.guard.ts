/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { OrganizationAccessGuard } from './organization-access.guard'
import { RequiredOrganizationResourcePermissions } from '../decorators/required-organization-resource-permissions.decorator'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationService } from '../services/organization.service'
import { OrganizationUserService } from '../services/organization-user.service'
import { SystemRole } from '../../user/enums/system-role.enum'
import { RequestWithOrganizationContext } from '../../common/types/request.types'

@Injectable()
export class OrganizationResourceActionGuard extends OrganizationAccessGuard {
  protected readonly logger = new Logger(OrganizationResourceActionGuard.name)

  constructor(
    organizationService: OrganizationService,
    organizationUserService: OrganizationUserService,
    private readonly reflector: Reflector,
  ) {
    super(organizationService, organizationUserService)
  }
  async canActivate(context: ExecutionContext): Promise<boolean> {
    const canActivate = await super.canActivate(context)

    const request = context.switchToHttp().getRequest<RequestWithOrganizationContext>()
    const authContext = request.user

    if (authContext.role === SystemRole.ADMIN) {
      return true
    }

    if (!canActivate) {
      return false
    }

    if (!authContext.organizationUser) {
      return false
    }

    if (authContext.organizationUser.role === OrganizationMemberRole.OWNER && !authContext.apiKey) {
      return true
    }

    const requiredPermissions =
      this.reflector.get(RequiredOrganizationResourcePermissions, context.getHandler()) ||
      this.reflector.get(RequiredOrganizationResourcePermissions, context.getClass())

    if (!requiredPermissions) {
      return true
    }

    const assignedPermissions = authContext.apiKey
      ? new Set(authContext.apiKey.permissions)
      : new Set(authContext.organizationUser.assignedRoles.flatMap((role) => role.permissions))

    return requiredPermissions.every((permission) => assignedPermissions.has(permission))
  }
}
