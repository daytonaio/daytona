/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationRegionController } from './organization-region.controller'
import { OrganizationAuthContextGuard } from '../guards/organization-auth-context.guard'
import { RegionAccessGuard } from '../../region/guards/region-access.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import {
  getAuthContextGuards,
  getResourceAccessGuards,
  getAllowedAuthStrategies,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] OrganizationRegionController', () => {
  const trackMethod = createCoverageTracker(OrganizationRegionController)

  it('listAvailableRegions', () => {
    const methodName = trackMethod('listAvailableRegions')
    expect(isPublicEndpoint(OrganizationRegionController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRegionController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(OrganizationRegionController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRegionController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationRegionController, methodName)).toBeUndefined()
  })

  it('createRegion', () => {
    const methodName = trackMethod('createRegion')
    expect(isPublicEndpoint(OrganizationRegionController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRegionController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(OrganizationRegionController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRegionController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(OrganizationRegionController, methodName), [
      OrganizationResourcePermission.WRITE_REGIONS,
    ])
  })

  it('getRegionById', () => {
    const methodName = trackMethod('getRegionById')
    expect(isPublicEndpoint(OrganizationRegionController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRegionController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(OrganizationRegionController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(OrganizationRegionController, methodName), [RegionAccessGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRegionController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(OrganizationRegionController, methodName)).toBeUndefined()
  })

  it('deleteRegion', () => {
    const methodName = trackMethod('deleteRegion')
    expect(isPublicEndpoint(OrganizationRegionController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRegionController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(OrganizationRegionController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(OrganizationRegionController, methodName), [RegionAccessGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRegionController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(OrganizationRegionController, methodName), [
      OrganizationResourcePermission.DELETE_REGIONS,
    ])
  })

  it('regenerateProxyApiKey', () => {
    const methodName = trackMethod('regenerateProxyApiKey')
    expect(isPublicEndpoint(OrganizationRegionController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRegionController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(OrganizationRegionController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(OrganizationRegionController, methodName), [RegionAccessGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRegionController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(OrganizationRegionController, methodName), [
      OrganizationResourcePermission.WRITE_REGIONS,
    ])
  })

  it('updateRegion', () => {
    const methodName = trackMethod('updateRegion')
    expect(isPublicEndpoint(OrganizationRegionController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRegionController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(OrganizationRegionController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(OrganizationRegionController, methodName), [RegionAccessGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRegionController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(OrganizationRegionController, methodName), [
      OrganizationResourcePermission.WRITE_REGIONS,
    ])
  })

  it('regenerateSshGatewayApiKey', () => {
    const methodName = trackMethod('regenerateSshGatewayApiKey')
    expect(isPublicEndpoint(OrganizationRegionController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRegionController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(OrganizationRegionController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(OrganizationRegionController, methodName), [RegionAccessGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRegionController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(OrganizationRegionController, methodName), [
      OrganizationResourcePermission.WRITE_REGIONS,
    ])
  })

  it('regenerateSnapshotManagerCredentials', () => {
    const methodName = trackMethod('regenerateSnapshotManagerCredentials')
    expect(isPublicEndpoint(OrganizationRegionController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(OrganizationRegionController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(OrganizationRegionController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(OrganizationRegionController, methodName), [RegionAccessGuard])
    expect(getRequiredOrganizationMemberRole(OrganizationRegionController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(OrganizationRegionController, methodName), [
      OrganizationResourcePermission.WRITE_REGIONS,
    ])
  })
})
