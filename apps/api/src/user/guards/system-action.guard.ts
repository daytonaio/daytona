/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { RequiredSystemRole } from '../decorators/required-system-role.decorator'
import { SystemRole } from '../enums/system-role.enum'
import { isBaseAuthContext } from '../../common/interfaces/auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { isPublic } from '../../auth/decorators/public.decorator'

/**
 * Authentication guard that enforces the `RequiredSystemRole` decorator.
 *
 * Access is granted if the user's role matches any of the required roles.
 * If no role requirement is set on the handler or controller, access is granted by default.
 */
@Injectable()
export class SystemActionGuard implements CanActivate {
  protected readonly logger = new Logger(SystemActionGuard.name)

  constructor(private readonly reflector: Reflector) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    if (isPublic(context, this.reflector)) {
      return true
    }

    const authContext = getAuthContext(context, isBaseAuthContext)

    let requiredRole: SystemRole | SystemRole[] =
      this.reflector.get(RequiredSystemRole, context.getHandler()) ||
      this.reflector.get(RequiredSystemRole, context.getClass())

    if (!requiredRole) {
      return true
    }

    if (!Array.isArray(requiredRole)) {
      requiredRole = [requiredRole]
    }

    return (requiredRole as string[]).includes(authContext.role as string)
  }
}
