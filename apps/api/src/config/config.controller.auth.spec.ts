/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConfigController } from './config.controller'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../test/helpers/controller-metadata.helper'

describe('[AUTH] ConfigController', () => {
  const trackMethod = createCoverageTracker(ConfigController)

  it('getConfig', () => {
    const methodName = trackMethod('getConfig')
    expect(isPublicEndpoint(ConfigController, methodName)).toBe(true)
    expectArrayMatch(getAllowedAuthStrategies(ConfigController, methodName), [])
    expectArrayMatch(getAuthContextGuards(ConfigController, methodName), [])
  })
})
