/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExecutionContext, UnauthorizedException } from '@nestjs/common'
import { IAuthContext } from '../common/interfaces/auth-context.interface'

export function getAuthContext<T extends IAuthContext>(
  context: ExecutionContext,
  isFunction: (user: IAuthContext) => user is T,
): T {
  const request = context.switchToHttp().getRequest()

  if (request.user && isFunction(request.user)) {
    return request.user
  }

  throw new UnauthorizedException('Unauthorized')
}
