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
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] OrganizationRegionController', () => {
  const trackMethod = createCoverageTracker(OrganizationRegionController)

  it('listAvailableRegions', () => {
    const methodName = trackMethod('listAvailableRegions')
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
