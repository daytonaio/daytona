/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotFoundException } from '@nestjs/common'
import { RunnerAccessGuard } from './runner-access.guard'
import {
  createMockHealthCheckAuthContext,
  createMockOtelCollectorAuthContext,
  createMockUserAuthContext,
  createMockProxyAuthContext,
  createMockRegionProxyAuthContext,
  createMockRegionSshGatewayAuthContext,
  createMockRunnerAuthContext,
  createMockSshGatewayAuthContext,
  createMockRegionAuthContext,
  createMockOrganizationAuthContext,
} from '../../test/helpers/auth-context.factory'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'

describe('[AUTH] RunnerAccessGuard', () => {
  let guard: RunnerAccessGuard
  const runnerService: any = { findOneOrFail: jest.fn() }
  const regionService: any = { findOne: jest.fn() }

  beforeEach(() => {
    guard = new RunnerAccessGuard(runnerService, regionService)
    runnerService.findOneOrFail.mockReset()
    regionService.findOne.mockReset()
  })

  it('allows RegionAuthContext with matching region', async () => {
    const authContext = createMockRegionAuthContext()
    const runner = { id: 'runner-1', region: authContext.regionId }
    runnerService.findOneOrFail.mockResolvedValue(runner)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: runner.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects RegionAuthContext with non-matching region', async () => {
    const authContext = createMockRegionAuthContext()
    const runner = { id: 'runner-1', region: 'other-region' }
    runnerService.findOneOrFail.mockResolvedValue(runner)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: runner.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it('allows ProxyAuthContext', async () => {
    const authContext = createMockProxyAuthContext()
    const runner = { id: 'runner-1', region: 'region-1' }
    runnerService.findOneOrFail.mockResolvedValue(runner)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: runner.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('allows SshGatewayAuthContext', async () => {
    const authContext = createMockSshGatewayAuthContext()
    const runner = { id: 'runner-1', region: 'region-1' }
    runnerService.findOneOrFail.mockResolvedValue(runner)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: runner.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('allows OrganizationAuthContext with matching region', async () => {
    const authContext = createMockOrganizationAuthContext()
    const runner = { id: 'runner-1', region: 'region-1' }
    runnerService.findOneOrFail.mockResolvedValue(runner)
    regionService.findOne.mockResolvedValue({
      id: 'region-1',
      organizationId: authContext.organizationId,
      regionType: 'custom',
    })
    const { context } = createMockExecutionContext({ user: authContext, params: { id: runner.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects OrganizationAuthContext with non-matching org', async () => {
    const authContext = createMockOrganizationAuthContext()
    const runner = { id: 'runner-1', region: 'region-1' }
    runnerService.findOneOrFail.mockResolvedValue(runner)
    regionService.findOne.mockResolvedValue({ id: 'region-1', organizationId: 'other-org', regionType: 'custom' })
    const { context } = createMockExecutionContext({ user: authContext, params: { id: runner.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it('rejects OrganizationAuthContext with matching org (if region not manually created by the organization)', async () => {
    const authContext = createMockOrganizationAuthContext()
    const runner = { id: 'runner-1', region: 'region-1' }
    runnerService.findOneOrFail.mockResolvedValue(runner)
    regionService.findOne.mockResolvedValue({
      id: 'region-1',
      organizationId: authContext.organizationId,
      regionType: 'internal',
    })
    const { context } = createMockExecutionContext({ user: authContext, params: { id: runner.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it.each([
    ['User', createMockUserAuthContext],
    ['Runner', createMockRunnerAuthContext],
    ['RegionProxy', createMockRegionProxyAuthContext],
    ['RegionSshGateway', createMockRegionSshGatewayAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('rejects %sAuthContext', async (_name, factory) => {
    const runner = { id: 'runner-1', region: 'region-1' }
    runnerService.findOneOrFail.mockResolvedValue(runner)
    const { context } = createMockExecutionContext({ user: factory(), params: { id: runner.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })
})
