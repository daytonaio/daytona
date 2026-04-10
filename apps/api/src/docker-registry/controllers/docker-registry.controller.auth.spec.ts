/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DockerRegistryController } from './docker-registry.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { DockerRegistryAccessGuard } from '../guards/docker-registry-access.guard'
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
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] DockerRegistryController', () => {
  const trackMethod = createCoverageTracker(DockerRegistryController)

  it('create', () => {
    const methodName = trackMethod('create')
    expectArrayMatch(getAllowedAuthStrategies(DockerRegistryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(DockerRegistryController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(DockerRegistryController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(DockerRegistryController, methodName), [
      OrganizationResourcePermission.WRITE_REGISTRIES,
    ])
  })

  it('findAll', () => {
    const methodName = trackMethod('findAll')
    expectArrayMatch(getAllowedAuthStrategies(DockerRegistryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(DockerRegistryController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(DockerRegistryController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(DockerRegistryController, methodName)).toBeUndefined()
  })

  it('getTransientPushAccess', () => {
    const methodName = trackMethod('getTransientPushAccess')
    expectArrayMatch(getAllowedAuthStrategies(DockerRegistryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(DockerRegistryController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(DockerRegistryController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(DockerRegistryController, methodName)).toBeUndefined()
  })

  it('findOne', () => {
    const methodName = trackMethod('findOne')
    expectArrayMatch(getAllowedAuthStrategies(DockerRegistryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(DockerRegistryController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(DockerRegistryController, methodName), [DockerRegistryAccessGuard])
    expect(getRequiredOrganizationMemberRole(DockerRegistryController, methodName)).toBeUndefined()
    expect(getRequiredOrganizationResourcePermissions(DockerRegistryController, methodName)).toBeUndefined()
  })

  it('update', () => {
    const methodName = trackMethod('update')
    expectArrayMatch(getAllowedAuthStrategies(DockerRegistryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(DockerRegistryController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(DockerRegistryController, methodName), [DockerRegistryAccessGuard])
    expect(getRequiredOrganizationMemberRole(DockerRegistryController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(DockerRegistryController, methodName), [
      OrganizationResourcePermission.WRITE_REGISTRIES,
    ])
  })

  it('remove', () => {
    const methodName = trackMethod('remove')
    expectArrayMatch(getAllowedAuthStrategies(DockerRegistryController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(DockerRegistryController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(DockerRegistryController, methodName), [DockerRegistryAccessGuard])
    expect(getRequiredOrganizationMemberRole(DockerRegistryController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(DockerRegistryController, methodName), [
      OrganizationResourcePermission.DELETE_REGISTRIES,
    ])
  })
})
