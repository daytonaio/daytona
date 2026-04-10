/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotFoundException } from '@nestjs/common'
import { DockerRegistryAccessGuard } from './docker-registry-access.guard'
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

describe('[AUTH] DockerRegistryAccessGuard', () => {
  let guard: DockerRegistryAccessGuard
  const registryService: any = { findOneOrFail: jest.fn() }

  beforeEach(() => {
    guard = new DockerRegistryAccessGuard(registryService)
    registryService.findOneOrFail.mockReset()
  })

  it('allows OrganizationAuthContext with matching org', async () => {
    const authContext = createMockOrganizationAuthContext()
    const registry = { id: 'reg-1', organizationId: authContext.organizationId, registryType: 'organization' }
    registryService.findOneOrFail.mockReturnValue(registry)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: registry.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects OrganizationAuthContext with non-matching org', async () => {
    const authContext = createMockOrganizationAuthContext()
    const registry = { id: 'reg-1', organizationId: 'other-org', registryType: 'organization' }
    registryService.findOneOrFail.mockReturnValue(registry)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: registry.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it('rejects OrganizationAuthContext with matching org (if registry not manually created by the organization)', async () => {
    const authContext = createMockOrganizationAuthContext()
    const registry = { id: 'reg-1', organizationId: authContext.organizationId, registryType: 'internal' }
    registryService.findOneOrFail.mockReturnValue(registry)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: registry.id } })
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
    const { context } = createMockExecutionContext({ user: factory(), params: { id: 'reg-1' } })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })
})
