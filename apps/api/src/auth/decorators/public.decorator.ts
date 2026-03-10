/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExecutionContext, SetMetadata } from '@nestjs/common'
import { Reflector } from '@nestjs/core'

export const IS_PUBLIC_KEY = 'isPublic'

/**
 * Marks a controller or handler as publicly accessible, bypassing authentication.
 *
 * Evaluated by all global guards.
 */
export const Public = () => SetMetadata(IS_PUBLIC_KEY, true)

/**
 * Returns `true` if the current handler or controller is decorated with `@Public()`.
 *
 * Use this in global guards to skip authentication/authorization for public endpoints.
 */
export function isPublic(context: ExecutionContext, reflector: Reflector): boolean {
  return reflector.getAllAndOverride<boolean>(IS_PUBLIC_KEY, [context.getHandler(), context.getClass()]) ?? false
}
