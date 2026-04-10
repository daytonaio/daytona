/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PreviewController } from './preview.controller'
import { ProxyAuthContextGuard } from '../guards/proxy-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  createCoverageTracker,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] PreviewController', () => {
  const trackMethod = createCoverageTracker(PreviewController)

  it('isSandboxPublic', () => {
    const methodName = trackMethod('isSandboxPublic')
    expectArrayMatch(getAllowedAuthStrategies(PreviewController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(PreviewController, methodName), [ProxyAuthContextGuard])
  })

  it('isValidAuthToken', () => {
    const methodName = trackMethod('isValidAuthToken')
    expectArrayMatch(getAllowedAuthStrategies(PreviewController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(PreviewController, methodName), [ProxyAuthContextGuard])
  })

  it('hasSandboxAccess', () => {
    const methodName = trackMethod('hasSandboxAccess')
    expectArrayMatch(getAllowedAuthStrategies(PreviewController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(PreviewController, methodName), [])
  })

  it('getSandboxIdFromSignedPreviewUrlToken', () => {
    const methodName = trackMethod('getSandboxIdFromSignedPreviewUrlToken')
    expectArrayMatch(getAllowedAuthStrategies(PreviewController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(PreviewController, methodName), [ProxyAuthContextGuard])
  })
})
