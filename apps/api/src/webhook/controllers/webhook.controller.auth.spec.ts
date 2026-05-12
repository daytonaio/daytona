/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { WebhookController } from './webhook.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] WebhookController', () => {
  const trackMethod = createCoverageTracker(WebhookController)

  it('getAppPortalAccess', () => {
    const methodName = trackMethod('getAppPortalAccess')
    expect(isPublicEndpoint(WebhookController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(WebhookController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(WebhookController, methodName), [OrganizationAuthContextGuard])
  })

  it('getInitializationStatus', () => {
    const methodName = trackMethod('getInitializationStatus')
    expect(isPublicEndpoint(WebhookController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(WebhookController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(WebhookController, methodName), [OrganizationAuthContextGuard])
  })

  it('refreshEndpoints', () => {
    const methodName = trackMethod('refreshEndpoints')
    expect(isPublicEndpoint(WebhookController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(WebhookController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(WebhookController, methodName), [OrganizationAuthContextGuard])
  })
})
