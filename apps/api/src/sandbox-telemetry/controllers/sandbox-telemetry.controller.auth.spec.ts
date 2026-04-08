/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxTelemetryController } from './sandbox-telemetry.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { SandboxAccessGuard } from '../../sandbox/guards/sandbox-access.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getResourceAccessGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  createCoverageTracker,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] SandboxTelemetryController', () => {
  const trackMethod = createCoverageTracker(SandboxTelemetryController)

  it('getSandboxLogs', () => {
    const methodName = trackMethod('getSandboxLogs')
    expectArrayMatch(getAllowedAuthStrategies(SandboxTelemetryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxTelemetryController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxTelemetryController, methodName), [SandboxAccessGuard])
  })

  it('getSandboxTraces', () => {
    const methodName = trackMethod('getSandboxTraces')
    expectArrayMatch(getAllowedAuthStrategies(SandboxTelemetryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxTelemetryController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxTelemetryController, methodName), [SandboxAccessGuard])
  })

  it('getSandboxTraceSpans', () => {
    const methodName = trackMethod('getSandboxTraceSpans')
    expectArrayMatch(getAllowedAuthStrategies(SandboxTelemetryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxTelemetryController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxTelemetryController, methodName), [SandboxAccessGuard])
  })

  it('getSandboxMetrics', () => {
    const methodName = trackMethod('getSandboxMetrics')
    expectArrayMatch(getAllowedAuthStrategies(SandboxTelemetryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SandboxTelemetryController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SandboxTelemetryController, methodName), [SandboxAccessGuard])
  })
})
