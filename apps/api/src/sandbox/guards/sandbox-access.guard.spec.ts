/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotFoundException } from '@nestjs/common'
import { SandboxAccessGuard } from './sandbox-access.guard'
import {
  createMockOrganizationAuthContext,
  createMockOtelCollectorAuthContext,
  createMockRunnerAuthContext,
  createMockProxyAuthContext,
  createMockSshGatewayAuthContext,
  createMockRegionAuthContext,
  createMockHealthCheckAuthContext,
  createMockUserAuthContext,
} from '../../test/helpers/auth-context.factory'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'

describe('[AUTH] SandboxAccessGuard', () => {
  let guard: SandboxAccessGuard
  const sandboxService: any = {
    getRunnerId: jest.fn(),
    getRegionId: jest.fn(),
    getOrganizationId: jest.fn(),
  }

  beforeEach(() => {
    guard = new SandboxAccessGuard(sandboxService)
    sandboxService.getRunnerId.mockReset()
    sandboxService.getRegionId.mockReset()
    sandboxService.getOrganizationId.mockReset()
  })

  it('allows RunnerAuthContext with matching runner', async () => {
    const authContext = createMockRunnerAuthContext()
    const sandbox = { id: 'sandbox-1', runnerId: authContext.runnerId }
    sandboxService.getRunnerId.mockReturnValue(sandbox.runnerId)
    const { context } = createMockExecutionContext({ user: authContext, params: { sandboxIdOrName: sandbox.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects RunnerAuthContext with non-matching runner', async () => {
    const authContext = createMockRunnerAuthContext()
    const sandbox = { id: 'sandbox-1', runnerId: 'different-runner' }
    sandboxService.getRunnerId.mockReturnValue(sandbox.runnerId)
    const { context } = createMockExecutionContext({ user: authContext, params: { sandboxIdOrName: sandbox.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it('allows RegionAuthContext with matching region', async () => {
    const authContext = createMockRegionAuthContext()
    const sandbox = { id: 'sandbox-1', regionId: authContext.regionId }
    sandboxService.getRegionId.mockReturnValue(sandbox.regionId)
    const { context } = createMockExecutionContext({ user: authContext, params: { sandboxIdOrName: sandbox.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects RegionAuthContext with non-matching region', async () => {
    const authContext = createMockRegionAuthContext()
    const sandbox = { id: 'sandbox-1', regionId: 'other-region' }
    sandboxService.getRegionId.mockReturnValue(sandbox.regionId)
    const { context } = createMockExecutionContext({ user: authContext, params: { sandboxIdOrName: sandbox.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it('allows ProxyAuthContext', async () => {
    const authContext = createMockProxyAuthContext()
    const sandbox = { id: 'sandbox-1' }
    const { context } = createMockExecutionContext({ user: authContext, params: { sandboxIdOrName: sandbox.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('allows SshGatewayAuthContext', async () => {
    const authContext = createMockSshGatewayAuthContext()
    const sandbox = { id: 'sandbox-1' }
    const { context } = createMockExecutionContext({ user: authContext, params: { sandboxIdOrName: sandbox.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('allows OrganizationAuthContext with matching organization', async () => {
    const authContext = createMockOrganizationAuthContext()
    const sandbox = { id: 'sandbox-1' }
    sandboxService.getOrganizationId.mockImplementation((id: string, orgId: string) => orgId)
    const { context } = createMockExecutionContext({ user: authContext, params: { sandboxIdOrName: sandbox.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects OrganizationAuthContext with non-matching organization', async () => {
    const authContext = createMockOrganizationAuthContext()
    const sandbox = { id: 'sandbox-1' }
    sandboxService.getOrganizationId.mockReturnValue('other-org')
    const { context } = createMockExecutionContext({ user: authContext, params: { sandboxIdOrName: sandbox.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it.each([
    ['User', createMockUserAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('rejects %sAuthContext', async (_name, factory) => {
    const { context } = createMockExecutionContext({ user: factory(), params: { sandboxIdOrName: 'sandbox-1' } } as any)
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })
})
