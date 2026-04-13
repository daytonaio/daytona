/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ObjectStorageController } from './object-storage.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] ObjectStorageController', () => {
  const trackMethod = createCoverageTracker(ObjectStorageController)

  it('getPushAccess', () => {
    const methodName = trackMethod('getPushAccess')
    expect(isPublicEndpoint(ObjectStorageController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(ObjectStorageController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(ObjectStorageController, methodName), [OrganizationAuthContextGuard])
  })
})
