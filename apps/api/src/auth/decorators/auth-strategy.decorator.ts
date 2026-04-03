/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Reflector } from '@nestjs/core'
import { AuthStrategyType } from '../enums/auth-strategy-type.enum'

/**
 * Restricts a controller or handler to specific authentication strategies.
 * When multiple strategies are provided, access is granted if the request uses _any_ of them.
 *
 * Evaluated by `GlobalAuthGuard`.
 */
export const AuthStrategy = Reflector.createDecorator<AuthStrategyType | AuthStrategyType[]>()
