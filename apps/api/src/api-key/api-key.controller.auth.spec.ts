/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKeyController } from './api-key.controller'
import { OrganizationAuthContextGuard } from '../organization/guards/organization-auth-context.guard'
import { AuthStrategyType } from '../auth/enums/auth-strategy-type.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  expectArrayMatch,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  createCoverageTracker,
  isPublicEndpoint,
} from '../test/helpers/controller-metadata.helper'

describe('[AUTH] ApiKeyController', () => {
  const trackMethod = createCoverageTracker(ApiKeyController)
  it('createApiKey', () => {
    const methodName = trackMethod('createApiKey')
    expect(isPublicEndpoint(ApiKeyController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('getApiKeys', () => {
    const methodName = trackMethod('getApiKeys')
    expect(isPublicEndpoint(ApiKeyController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('getCurrentApiKey', () => {
    const methodName = trackMethod('getCurrentApiKey')
    expect(isPublicEndpoint(ApiKeyController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('getApiKey', () => {
    const methodName = trackMethod('getApiKey')
    expect(isPublicEndpoint(ApiKeyController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('deleteApiKey', () => {
    const methodName = trackMethod('deleteApiKey')
    expect(isPublicEndpoint(ApiKeyController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('deleteApiKeyForUser', () => {
    const methodName = trackMethod('deleteApiKeyForUser')
    expect(isPublicEndpoint(ApiKeyController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })
})
