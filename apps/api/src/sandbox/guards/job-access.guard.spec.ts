/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotFoundException } from '@nestjs/common'
import { JobAccessGuard } from './job-access.guard'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'
import {
  createMockHealthCheckAuthContext,
  createMockOtelCollectorAuthContext,
  createMockOrganizationAuthContext,
  createMockProxyAuthContext,
  createMockRegionProxyAuthContext,
  createMockRegionSshGatewayAuthContext,
  createMockRunnerAuthContext,
  createMockSshGatewayAuthContext,
  createMockUserAuthContext,
} from '../../test/helpers/auth-context.factory'
import { createMockExecutionContext } from '../../test/helpers/execution-context.factory'

describe('[AUTH] JobAccessGuard', () => {
  let guard: JobAccessGuard
  const jobService: any = { findOne: jest.fn() }

  beforeEach(() => {
    guard = new JobAccessGuard(jobService)
    jobService.findOne.mockReset()
  })

  it('allows RunnerAuthContext with matching runnerId', async () => {
    const authContext = createMockRunnerAuthContext()
    const job = { id: 'job-1', runnerId: authContext.runnerId }
    jobService.findOne.mockReturnValue(job)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: job.id } })
    await expect(guard.canActivate(context)).resolves.toBe(true)
  })

  it('rejects RunnerAuthContext with non-matching runner', async () => {
    const authContext = createMockRunnerAuthContext()
    const job = { id: 'job-1', runnerId: 'other' }
    jobService.findOne.mockReturnValue(job)
    const { context } = createMockExecutionContext({ user: authContext, params: { id: job.id } })
    await expect(guard.canActivate(context)).rejects.toThrow(NotFoundException)
  })

  it.each([
    ['User', createMockUserAuthContext],
    ['Organization', createMockOrganizationAuthContext],
    ['Proxy', createMockProxyAuthContext],
    ['SshGateway', createMockSshGatewayAuthContext],
    ['RegionProxy', createMockRegionProxyAuthContext],
    ['RegionSshGateway', createMockRegionSshGatewayAuthContext],
    ['HealthCheck', createMockHealthCheckAuthContext],
    ['OtelCollector', createMockOtelCollectorAuthContext],
  ])('rejects %sAuthContext', async (_name, factory) => {
    const { context } = createMockExecutionContext({ user: factory(), params: { id: 'job-1' } })
    await expect(guard.canActivate(context)).rejects.toThrow(InvalidAuthenticationContextException)
  })
})
