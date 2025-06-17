/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { RequiredSystemRole, RequiredApiRole } from '../common/decorators/required-role.decorator'
import { SystemRole } from '../user/enums/system-role.enum'
import { ApiRole, AuthContext } from '../common/interfaces/auth-context.interface'

@Injectable()
export class SystemActionGuard implements CanActivate {
  protected readonly logger = new Logger(SystemActionGuard.name)

  constructor(private readonly reflector: Reflector) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    // TODO: initialize authContext safely
    const authContext: AuthContext = request.user

    let requiredRole: SystemRole | SystemRole[] | ApiRole | ApiRole[] =
      this.reflector.get(RequiredSystemRole, context.getHandler()) ||
      this.reflector.get(RequiredSystemRole, context.getClass())

    if (!requiredRole) {
      requiredRole =
        this.reflector.get(RequiredApiRole, context.getHandler()) ||
        this.reflector.get(RequiredApiRole, context.getClass())
      if (!requiredRole) {
        return true
      }
    }

    if (!Array.isArray(requiredRole)) {
      requiredRole = [requiredRole]
    }

    return (requiredRole as string[]).includes(authContext.role as string)
  }
}
