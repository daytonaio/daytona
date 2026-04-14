/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RegionController } from './region.controller'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] RegionController', () => {
  const trackMethod = createCoverageTracker(RegionController)

  it('listRegions', () => {
    const methodName = trackMethod('listRegions')
    expect(isPublicEndpoint(RegionController, methodName)).toBe(true)
    expectArrayMatch(getAllowedAuthStrategies(RegionController, methodName), [])
    expectArrayMatch(getAuthContextGuards(RegionController, methodName), [])
  })
})
