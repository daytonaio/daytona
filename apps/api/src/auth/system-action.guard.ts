/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { RequiredSystemRole } from '../common/decorators/required-system-role.decorator'
import { SystemRole } from '../user/enums/system-role.enum'
import { AuthContext } from '../common/interfaces/auth-context.interface'

@Injectable()
export class SystemActionGuard implements CanActivate {
  protected readonly logger = new Logger(SystemActionGuard.name)

  constructor(private readonly reflector: Reflector) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    // TODO: initialize authContext safely
    const authContext: AuthContext = request.user

    const requiredRole =
      this.reflector.get(RequiredSystemRole, context.getHandler()) ||
      this.reflector.get(RequiredSystemRole, context.getClass())

    if (!requiredRole) {
      return true
    }

    if (requiredRole === SystemRole.ADMIN) {
      return authContext.role === SystemRole.ADMIN
    }

    return true
  }
}
