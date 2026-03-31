/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExecutionContext } from '@nestjs/common'
import { BaseAuthContext } from '../interfaces/base-auth-context.interface'
import { InvalidAuthenticationContextException } from '../exceptions/invalid-authentication-context.exception'

/**
 * Retrieves the authentication context from the request.
 *
 * @throws {InvalidAuthenticationContextException} if the context is not found or does not match the expected type.
 */
export function getAuthContext<T extends BaseAuthContext>(
  context: ExecutionContext,
  isFunction: (user: unknown) => user is T,
): T {
  const request = context.switchToHttp().getRequest()

  if (request.user && isFunction(request.user)) {
    return request.user
  }

  throw new InvalidAuthenticationContextException()
}
