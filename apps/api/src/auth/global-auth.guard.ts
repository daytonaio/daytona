/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExecutionContext, Injectable, Logger, UnauthorizedException } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { AuthGuard } from '@nestjs/passport'
import { AuthStrategyType } from './enums/auth-strategy-type.enum'
import { RequestWithAuthMetadata } from './interfaces/request-with-auth-metadata.interface'
import { isPublic } from './decorators/public.decorator'
import { AuthStrategy } from './decorators/auth-strategy.decorator'

/**
 * Global authentication guard for the application.
 */
@Injectable()
export class GlobalAuthGuard extends AuthGuard([AuthStrategyType.API_KEY, AuthStrategyType.JWT]) {
  private readonly logger = new Logger(GlobalAuthGuard.name)

  constructor(private readonly reflector: Reflector) {
    super()
  }

  /**
   * Runs each registered strategy in order until one succeeds or all fail.
   *
   * Endpoints decorated with `@Public()` bypass authentication entirely.
   */
  canActivate(context: ExecutionContext) {
    if (isPublic(context, this.reflector)) {
      return true
    }

    const request = context.switchToHttp().getRequest<RequestWithAuthMetadata>()
    const allowedStrategies = this.getAllowedStrategies(context)
    request.authMetadata = {
      isStrategyAllowed: (type) => allowedStrategies.includes(type),
    }

    return super.canActivate(context)
  }

  /**
   * Invoked once after a strategy succeeds or all allowed strategies fail.
   *
   * It returns the authenticated user object or throws a generic `UnauthorizedException`.
   */
  handleRequest(err: any, user: any, info: any) {
    if (err || !user) {
      this.logger.debug('Authentication failed', { err, user })
      throw new UnauthorizedException('Invalid credentials')
    }

    return user
  }

  /**
   * Gets the allowed strategies for the current execution context.
   *
   * Defaults to JWT-only when no `@AuthStrategy()` decorator is present.
   */
  private getAllowedStrategies(context: ExecutionContext): AuthStrategyType[] {
    const value = this.reflector.getAllAndOverride<AuthStrategyType | AuthStrategyType[]>(AuthStrategy, [
      context.getHandler(),
      context.getClass(),
    ])

    if (!value) {
      return [AuthStrategyType.JWT]
    }

    return Array.isArray(value) ? value : [value]
  }
}
