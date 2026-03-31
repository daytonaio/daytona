/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createParamDecorator, ExecutionContext } from '@nestjs/common'
import { BaseAuthContext, isBaseAuthContext } from '../interfaces/base-auth-context.interface'
import { isUserAuthContext } from '../interfaces/user-auth-context.interface'
import { isOrganizationAuthContext } from '../interfaces/organization-auth-context.interface'
import { isRunnerAuthContext } from '../interfaces/runner-auth-context.interface'
import { getAuthContext } from '../utils/get-auth-context'

/**
 * Parameter decorator that extracts and validates the authenticated user context from the request.
 *
 * Accepts a type guard to validate the context type at runtime.
 */
export const AuthContext = createParamDecorator(
  <T extends BaseAuthContext>(isFunction: (user: unknown) => user is T, ctx: ExecutionContext): T => {
    return getAuthContext(ctx, isFunction)
  },
)

/**
 * Shorthand for `@AuthContext(isBaseAuthContext)`.
 *
 * Extracts the authenticated user context and validates it is a {@link BaseAuthContext} at runtime.
 */
export const IsBaseAuthContext = () => AuthContext(isBaseAuthContext)

/**
 * Shorthand for `@AuthContext(isUserAuthContext)`.
 *
 * Extracts the authenticated user context and validates it is a {@link UserAuthContext} at runtime.
 */
export const IsUserAuthContext = () => AuthContext(isUserAuthContext)

/**
 * Shorthand for `@AuthContext(isOrganizationAuthContext)`.
 *
 * Extracts the authenticated user context and validates it is an {@link OrganizationAuthContext} at runtime.
 */
export const IsOrganizationAuthContext = () => AuthContext(isOrganizationAuthContext)

/**
 * Shorthand for `@AuthContext(isRunnerAuthContext)`.
 *
 * Extracts the authenticated user context and validates it is a {@link RunnerAuthContext} at runtime.
 */
export const IsRunnerAuthContext = () => AuthContext(isRunnerAuthContext)
