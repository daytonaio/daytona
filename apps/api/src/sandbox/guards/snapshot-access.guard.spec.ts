/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotFoundException } from '@nestjs/common'
import { SnapshotAccessGuard } from './snapshot-access.guard'
import {
  createMockHealthCheckAuthContext,
  createMockOtelCollectorAuthContext,
  createMockProxyAuthContext,
  createMockRegionProxyAuthContext,
  createMockRegionSshGatewayAuthContext,
  createMockRunnerAuthContext,
  createMockSshGatewayAuthContext,
  createMockRegionAuthContext,
  createMockOrganizationAuthContext,
  createMockUserAuthContext,
} from '../../test/helpers/auth-context.factory'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'

describe('[AUTH] SnapshotAccessGuard', () => {
  let guard: SnapshotAccessGuard
  const snapshotService: any = {
    getSnapshot: jest.fn(),
    getSnapshotByName: jest.fn(),
    isAvailableInRegion: jest.fn(),
  }

  beforeEach(() => {
    guard = new SnapshotAccessGuard(snapshotService)
    snapshotService.getSnapshot.mockReset()
    snapshotService.getSnapshotByName.mockReset()
    snapshotService.isAvailableInRegion.mockReset()
  })

  it('allows RegionAuthContext when snapshot available in region', async () => {
    const authContext = createMockRegionAuthContext()
    const snapshot = { id: 'snap-1', regionId: authContext.regionId }
    snapshotService.getSnapshot.mockReturnValue(snapshot)
    snapshotService.isAvailableInRegion.mockReturnValue(true)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: snapshot.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects RegionAuthContext when snapshot not available in region', async () => {
    const authContext = createMockRegionAuthContext()
    const snapshot = { id: 'snap-1', regionId: authContext.regionId }
    snapshotService.getSnapshot.mockReturnValue(snapshot)
    snapshotService.isAvailableInRegion.mockReturnValue(false)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: snapshot.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it('allows ProxyAuthContext', async () => {
    const authContext = createMockProxyAuthContext()
    const snapshot = { id: 'snap-1' }
    const { context } = createMockExecutionContext({ user: authContext, params: { id: snapshot.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('allows SshGatewayAuthContext', async () => {
    const authContext = createMockSshGatewayAuthContext()
    const snapshot = { id: 'snap-1' }
    const { context } = createMockExecutionContext({ user: authContext, params: { id: snapshot.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('allows OrganizationAuthContext with matching organization', async () => {
    const authContext = createMockOrganizationAuthContext()
    const snapshot = { id: 'snap-1', organizationId: authContext.organizationId, regionId: 'r1' }
    snapshotService.getSnapshot.mockReturnValue(snapshot)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: snapshot.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects OrganizationAuthContext with non-matching organization', async () => {
    const authContext = createMockOrganizationAuthContext()
    const snapshot = { id: 'snap-1', organizationId: 'other-org', regionId: 'r1' }
    snapshotService.getSnapshot.mockReturnValue(snapshot)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: snapshot.id } })
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
    const snapshot = { id: 'snap-1', organizationId: 'org-1', regionId: 'r1' }
    snapshotService.getSnapshot.mockReturnValue(snapshot)
    const { context } = createMockExecutionContext({ user: factory(), params: { id: snapshot.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })
})
