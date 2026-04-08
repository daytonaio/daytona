/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotFoundException } from '@nestjs/common'
import { RegionAccessGuard } from './region-access.guard'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'
import {
  createMockHealthCheckAuthContext,
  createMockOrganizationAuthContext,
  createMockOtelCollectorAuthContext,
  createMockProxyAuthContext,
  createMockRegionProxyAuthContext,
  createMockRegionSshGatewayAuthContext,
  createMockRunnerAuthContext,
  createMockSshGatewayAuthContext,
  createMockUserAuthContext,
} from '../../test/helpers/auth-context.factory'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'

describe('[AUTH] RegionAccessGuard', () => {
  let guard: RegionAccessGuard
  const regionService: any = { findOne: jest.fn() }

  beforeEach(() => {
    guard = new RegionAccessGuard(regionService)
    regionService.findOne.mockReset()
  })

  it('allows OrganizationAuthContext with matching region', async () => {
    const authContext = createMockOrganizationAuthContext()
    const region = { id: 'region-1', organizationId: authContext.organizationId, regionType: 'custom' }
    regionService.findOne.mockReturnValue(region)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: region.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects OrganizationAuthContext with non-matching org', async () => {
    const authContext = createMockOrganizationAuthContext()
    const region = { id: 'region-1', organizationId: 'other-org', regionType: 'custom' }
    regionService.findOne.mockReturnValue(region)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: region.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it('rejects OrganizationAuthContext with matching org (if region not manually created by the organization)', async () => {
    const authContext = createMockOrganizationAuthContext()
    const region = { id: 'region-1', organizationId: authContext.organizationId, regionType: 'dedicated' }
    regionService.findOne.mockReturnValue(region)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: region.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it.each([
    ['User', createMockUserAuthContext],
    ['Runner', createMockRunnerAuthContext],
    ['Proxy', createMockProxyAuthContext],
    ['SshGateway', createMockSshGatewayAuthContext],
    ['RegionProxy', createMockRegionProxyAuthContext],
    ['RegionSshGateway', createMockRegionSshGatewayAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('rejects %sAuthContext', async (_name, factory) => {
    const { context } = createMockExecutionContext({ user: factory(), params: { id: 'region-1' } })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })
})
