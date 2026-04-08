/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { HealthController } from './health.controller'
import { HealthCheckAuthContextGuard } from './guards/health-check-auth-context.guard'
import { AuthStrategyType } from '../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  createCoverageTracker,
} from '../test/helpers/controller-metadata.helper'

describe('[AUTH] HealthController', () => {
  const trackMethod = createCoverageTracker(HealthController)

  it('live', () => {
    const methodName = trackMethod('live')
    expectArrayMatch(getAllowedAuthStrategies(HealthController, methodName), [])
    expectArrayMatch(getAuthContextGuards(HealthController, methodName), [])
  })

  it('check', () => {
    const methodName = trackMethod('check')
    expectArrayMatch(getAllowedAuthStrategies(HealthController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(HealthController, methodName), [HealthCheckAuthContextGuard])
  })
})
