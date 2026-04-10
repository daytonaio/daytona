/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotFoundException } from '@nestjs/common'
import { VolumeAccessGuard } from './volume-access.guard'
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

describe('[AUTH] VolumeAccessGuard', () => {
  let guard: VolumeAccessGuard
  const volumeService: any = { getOrganizationId: jest.fn() }

  beforeEach(() => {
    guard = new VolumeAccessGuard(volumeService)
    volumeService.getOrganizationId.mockReset()
  })

  it('allows OrganizationAuthContext with matching organization', async () => {
    const authContext = createMockOrganizationAuthContext()
    const volume = { id: 'vol-1', organizationId: authContext.organizationId }
    volumeService.getOrganizationId.mockReturnValue(volume.organizationId)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: volume.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects OrganizationAuthContext with non-matching organization', async () => {
    const authContext = createMockOrganizationAuthContext()
    const volume = { id: 'vol-1', organizationId: 'other-org' }
    volumeService.getOrganizationId.mockReturnValue(volume.organizationId)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: volume.id } })
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
    const { context } = createMockExecutionContext({ user: factory(), params: { id: 'vol-1' } } as any)
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })
})
