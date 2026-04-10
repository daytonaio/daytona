/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { UserController } from './user.controller'
import { AuthStrategyType } from '../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  createCoverageTracker,
} from '../test/helpers/controller-metadata.helper'
import { UserAuthContextGuard } from './guards/user-auth-context.guard'

describe('[AUTH] UserController', () => {
  const trackMethod = createCoverageTracker(UserController)

  it('getAuthenticatedUser', () => {
    const methodName = trackMethod('getAuthenticatedUser')

    expectArrayMatch(getAllowedAuthStrategies(UserController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(UserController, methodName), [UserAuthContextGuard])
  })

  it('getAvailableAccountProviders', () => {
    const methodName = trackMethod('getAvailableAccountProviders')
    expectArrayMatch(getAllowedAuthStrategies(UserController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(UserController, methodName), [UserAuthContextGuard])
  })

  it('linkAccount', () => {
    const methodName = trackMethod('linkAccount')
    expectArrayMatch(getAllowedAuthStrategies(UserController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(UserController, methodName), [UserAuthContextGuard])
  })

  it('unlinkAccount', () => {
    const methodName = trackMethod('unlinkAccount')
    expectArrayMatch(getAllowedAuthStrategies(UserController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(UserController, methodName), [UserAuthContextGuard])
  })

  it('enrollInSmsMfa', () => {
    const methodName = trackMethod('enrollInSmsMfa')
    expectArrayMatch(getAllowedAuthStrategies(UserController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(UserController, methodName), [UserAuthContextGuard])
  })
})
