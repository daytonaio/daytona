/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExecutionContext, UnauthorizedException } from '@nestjs/common'
import { BaseAuthContext } from '../common/interfaces/auth-context.interface'

/**
 * Retrieves the authentication context from the request.
 *
 * Throws a generic `UnauthorizedException` if the context is not found or does not match the expected type.
 */
export function getAuthContext<T extends BaseAuthContext>(
  context: ExecutionContext,
  isFunction: (user: BaseAuthContext) => user is T,
): T {
  const request = context.switchToHttp().getRequest()

  if (request.user && isFunction(request.user)) {
    return request.user
  }

  throw new UnauthorizedException('Invalid authentication context')
}
