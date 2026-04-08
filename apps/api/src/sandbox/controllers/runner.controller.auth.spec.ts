/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RunnerController } from './runner.controller'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
import { RunnerAuthContextGuard } from '../guards/runner-auth-context.guard'
import { RunnerAccessGuard } from '../guards/runner-access.guard'
import { SandboxAccessGuard } from '../guards/sandbox-access.guard'
import { ProxyAuthContextGuard } from '../guards/proxy-auth-context.guard'
import { SshGatewayAuthContextGuard } from '../guards/ssh-gateway-auth-context.guard'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'
import {
  getAuthContextGuards,
  getResourceAccessGuards,
  getAllowedAuthStrategies,
  getRequiredOrganizationMemberRole,
  getRequiredOrganizationResourcePermissions,
  expectArrayMatch,
  createCoverageTracker,
} from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] RunnerController', () => {
  const trackMethod = createCoverageTracker(RunnerController)

  it('create', () => {
    const methodName = trackMethod('create')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(RunnerController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(RunnerController, methodName), [
      OrganizationResourcePermission.WRITE_RUNNERS,
    ])
  })

  it('getInfoForAuthenticatedRunner', () => {
    const methodName = trackMethod('getInfoForAuthenticatedRunner')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [RunnerAuthContextGuard])
  })

  it('getRunnerBySandboxId', () => {
    const methodName = trackMethod('getRunnerBySandboxId')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [
      ProxyAuthContextGuard,
      SshGatewayAuthContextGuard,
    ])
    expectArrayMatch(getResourceAccessGuards(RunnerController, methodName), [SandboxAccessGuard])
  })

  it('getRunnersBySnapshotRef', () => {
    const methodName = trackMethod('getRunnersBySnapshotRef')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [
      ProxyAuthContextGuard,
      SshGatewayAuthContextGuard,
    ])
  })

  it('getRunnerById', () => {
    const methodName = trackMethod('getRunnerById')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(RunnerController, methodName), [RunnerAccessGuard])
    expect(getRequiredOrganizationMemberRole(RunnerController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(RunnerController, methodName), [
      OrganizationResourcePermission.READ_RUNNERS,
    ])
  })

  it('getRunnerByIdFull', () => {
    const methodName = trackMethod('getRunnerByIdFull')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [
      ProxyAuthContextGuard,
      SshGatewayAuthContextGuard,
    ])
    expectArrayMatch(getResourceAccessGuards(RunnerController, methodName), [RunnerAccessGuard])
  })

  it('findAll', () => {
    const methodName = trackMethod('findAll')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [OrganizationAuthContextGuard])
    expect(getRequiredOrganizationMemberRole(RunnerController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(RunnerController, methodName), [
      OrganizationResourcePermission.READ_RUNNERS,
    ])
  })

  it('updateSchedulingStatus', () => {
    const methodName = trackMethod('updateSchedulingStatus')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(RunnerController, methodName), [RunnerAccessGuard])
    expect(getRequiredOrganizationMemberRole(RunnerController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(RunnerController, methodName), [
      OrganizationResourcePermission.WRITE_RUNNERS,
    ])
  })

  it('updateDrainingStatus', () => {
    const methodName = trackMethod('updateDrainingStatus')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(RunnerController, methodName), [RunnerAccessGuard])
    expect(getRequiredOrganizationMemberRole(RunnerController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(RunnerController, methodName), [
      OrganizationResourcePermission.WRITE_RUNNERS,
    ])
  })

  it('delete', () => {
    const methodName = trackMethod('delete')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [
      AuthStrategyType.API_KEY,
      AuthStrategyType.JWT,
    ])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [OrganizationAuthContextGuard])
    expectArrayMatch(getResourceAccessGuards(RunnerController, methodName), [RunnerAccessGuard])
    expect(getRequiredOrganizationMemberRole(RunnerController, methodName)).toBeUndefined()
    expectArrayMatch(getRequiredOrganizationResourcePermissions(RunnerController, methodName), [
      OrganizationResourcePermission.DELETE_RUNNERS,
    ])
  })

  it('runnerHealthcheck', () => {
    const methodName = trackMethod('runnerHealthcheck')
    expectArrayMatch(getAllowedAuthStrategies(RunnerController, methodName), [AuthStrategyType.API_KEY])
    expectArrayMatch(getAuthContextGuards(RunnerController, methodName), [RunnerAuthContextGuard])
  })
})
