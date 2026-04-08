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
} from '../test/helpers/controller-metadata.helper'

describe('[AUTH] ApiKeyController', () => {
  const trackMethod = createCoverageTracker(ApiKeyController)
  it('createApiKey', () => {
    const methodName = trackMethod('createApiKey')
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('getApiKeys', () => {
    const methodName = trackMethod('getApiKeys')
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('getCurrentApiKey', () => {
    const methodName = trackMethod('getCurrentApiKey')
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('getApiKey', () => {
    const methodName = trackMethod('getApiKey')
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('deleteApiKey', () => {
    const methodName = trackMethod('deleteApiKey')
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })

  it('deleteApiKeyForUser', () => {
    const methodName = trackMethod('deleteApiKeyForUser')
    expectArrayMatch(getAllowedAuthStrategies(ApiKeyController, methodName), [AuthStrategyType.JWT])
    expectArrayMatch(getAuthContextGuards(ApiKeyController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(ApiKeyController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(ApiKeyController, methodName)).toBeUndefined()
  })
})
