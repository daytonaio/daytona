/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { VolumeController } from './volume.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { VolumeAccessGuard } from '../guards/volume-access.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import {
  getAuthContextGuards,
  getAllowedAuthStrategies,
  getResourceAccessGuards,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
  createCoverageTracker,
  isPublicEndpoint,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] VolumeController', () => {
  const trackMethod = createCoverageTracker(VolumeController)

  it('listVolumes', () => {
    const methodName = trackMethod('listVolumes')
    expect(isPublicEndpoint(VolumeController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(VolumeController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(VolumeController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(VolumeController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(VolumeController, methodName), [
      OrganizationResourcePermission.READ_VOLUMES,
    ])
  })

  it('createVolume', () => {
    const methodName = trackMethod('createVolume')
    expect(isPublicEndpoint(VolumeController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(VolumeController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(VolumeController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(VolumeController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(VolumeController, methodName), [
      OrganizationResourcePermission.WRITE_VOLUMES,
    ])
  })

  it('getVolume', () => {
    const methodName = trackMethod('getVolume')
    expect(isPublicEndpoint(VolumeController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(VolumeController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(VolumeController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(VolumeController, methodName), [VolumeAccessGuard])
    expect(getRequiredOrganizationMemberRole(VolumeController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(VolumeController, methodName), [
      OrganizationResourcePermission.READ_VOLUMES,
    ])
  })

  it('deleteVolume', () => {
    const methodName = trackMethod('deleteVolume')
    expect(isPublicEndpoint(VolumeController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(VolumeController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(VolumeController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(VolumeController, methodName), [VolumeAccessGuard])
    expect(getRequiredOrganizationMemberRole(VolumeController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(VolumeController, methodName), [
      OrganizationResourcePermission.DELETE_VOLUMES,
    ])
  })

  it('getVolumeByName', () => {
    const methodName = trackMethod('getVolumeByName')
    expect(isPublicEndpoint(VolumeController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(VolumeController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(VolumeController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(VolumeController, methodName), [VolumeAccessGuard])
    expect(getRequiredOrganizationMemberRole(VolumeController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(VolumeController, methodName), [
      OrganizationResourcePermission.READ_VOLUMES,
    ])
  })
})
