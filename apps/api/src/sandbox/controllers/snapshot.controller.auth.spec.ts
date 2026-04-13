/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotController } from './snapshot.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { SnapshotAccessGuard } from '../guards/snapshot-access.guard'
import { SnapshotReadAccessGuard } from '../guards/snapshot-read-access.guard'
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

describe('[AUTH] SnapshotController', () => {
  const trackMethod = createCoverageTracker(SnapshotController)

  it('createSnapshot', () => {
    const methodName = trackMethod('createSnapshot')
    expect(isPublicEndpoint(SnapshotController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SnapshotController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SnapshotController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(SnapshotController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SnapshotController, methodName), [
      OrganizationResourcePermission.WRITE_SNAPSHOTS,
    ])
  })

  it('getSnapshot', () => {
    const methodName = trackMethod('getSnapshot')
    expect(isPublicEndpoint(SnapshotController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SnapshotController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SnapshotController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SnapshotController, methodName), [SnapshotReadAccessGuard])
    expect(getRequiredOrganizationMemberRole(SnapshotController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SnapshotController, methodName)).toBeUndefined()
  })

  it('removeSnapshot', () => {
    const methodName = trackMethod('removeSnapshot')
    expect(isPublicEndpoint(SnapshotController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SnapshotController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SnapshotController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SnapshotController, methodName), [SnapshotAccessGuard])
    expect(getRequiredOrganizationMemberRole(SnapshotController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SnapshotController, methodName), [
      OrganizationResourcePermission.DELETE_SNAPSHOTS,
    ])
  })

  it('getAllSnapshots', () => {
    const methodName = trackMethod('getAllSnapshots')
    expect(isPublicEndpoint(SnapshotController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SnapshotController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SnapshotController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(SnapshotController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SnapshotController, methodName)).toBeUndefined()
  })

  it('getSnapshotBuildLogs', () => {
    const methodName = trackMethod('getSnapshotBuildLogs')
    expect(isPublicEndpoint(SnapshotController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SnapshotController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SnapshotController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SnapshotController, methodName), [SnapshotAccessGuard])
    expect(getRequiredOrganizationMemberRole(SnapshotController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SnapshotController, methodName)).toBeUndefined()
  })

  it('getSnapshotBuildLogsUrl', () => {
    const methodName = trackMethod('getSnapshotBuildLogsUrl')
    expect(isPublicEndpoint(SnapshotController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SnapshotController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SnapshotController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SnapshotController, methodName), [SnapshotAccessGuard])
    expect(getRequiredOrganizationMemberRole(SnapshotController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(SnapshotController, methodName)).toBeUndefined()
  })

  it('activateSnapshot', () => {
    const methodName = trackMethod('activateSnapshot')
    expect(isPublicEndpoint(SnapshotController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SnapshotController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SnapshotController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SnapshotController, methodName), [SnapshotAccessGuard])
    expect(getRequiredOrganizationMemberRole(SnapshotController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SnapshotController, methodName), [
      OrganizationResourcePermission.WRITE_SNAPSHOTS,
    ])
  })

  it('deactivateSnapshot', () => {
    const methodName = trackMethod('deactivateSnapshot')
    expect(isPublicEndpoint(SnapshotController, methodName)).toBe(false)
    expectArrayMatch(getAllowedAuthStrategies(SnapshotController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(SnapshotController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(SnapshotController, methodName), [SnapshotAccessGuard])
    expect(getRequiredOrganizationMemberRole(SnapshotController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(SnapshotController, methodName), [
      OrganizationResourcePermission.WRITE_SNAPSHOTS,
    ])
  })
})
