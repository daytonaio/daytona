/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Organization } from '../../organization/entities/organization.entity'
import { Runner } from '../entities/runner.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxService } from './sandbox.service'

interface StubDeps {
  sandboxRepository: {
    updateWhere: jest.Mock
  }
  runnerService: {
    findOneOrFail: jest.Mock
  }
  organizationService: {
    assertOrganizationIsNotSuspended: jest.Mock
  }
  runnerAdapterFactory: {
    create: jest.Mock
  }
  regionService: {
    findOne: jest.Mock
  }
  snapshotService: {
    validateOrganizationQuotas: jest.Mock
    createPendingSnapshotFromSandbox: jest.Mock
    failSnapshotFromSandbox: jest.Mock
    rollbackPendingUsage: jest.Mock
  }
  dockerRegistryService: {
    getAvailableInternalRegistry: jest.Mock
  }
}

const organization = { id: 'org-1' } as Organization
const runner = { id: 'runner-1', apiVersion: '2' } as Runner

function makeSandbox(): Sandbox {
  return {
    id: 'sandbox-1',
    organizationId: 'org-1',
    state: SandboxState.STARTED,
    pending: false,
    runnerId: 'runner-1',
    region: 'region-1',
    sandboxClass: SandboxClass.CONTAINER,
    cpu: 2,
    gpu: 0,
    mem: 4,
    disk: 10,
  } as Sandbox
}

function createService(): { service: SandboxService; stub: StubDeps } {
  const stub: StubDeps = {
    sandboxRepository: {
      updateWhere: jest.fn(async (_id: string, params: { updateData: Partial<Sandbox> }) => ({
        ...makeSandbox(),
        ...params.updateData,
      })),
    },
    runnerService: {
      findOneOrFail: jest.fn(async () => runner),
    },
    organizationService: {
      assertOrganizationIsNotSuspended: jest.fn(),
    },
    runnerAdapterFactory: {
      create: jest.fn(),
    },
    regionService: {
      findOne: jest.fn(async () => ({ id: 'region-1' })),
    },
    snapshotService: {
      validateOrganizationQuotas: jest.fn(async () => ({ pendingSnapshotCountIncremented: true })),
      createPendingSnapshotFromSandbox: jest.fn(async () => ({})),
      failSnapshotFromSandbox: jest.fn(async () => undefined),
      rollbackPendingUsage: jest.fn(async () => undefined),
    },
    dockerRegistryService: {
      getAvailableInternalRegistry: jest.fn(async () => ({ id: 'registry-1' })),
    },
  }

  // Test-only construction: createSnapshotFromSandbox exercises a small
  // subset of the 22 constructor dependencies; the rest stay undefined.
  const ServiceCtor = SandboxService as unknown as new (...args: unknown[]) => SandboxService
  const service = new ServiceCtor(
    stub.sandboxRepository, // sandboxRepository
    undefined, // snapshotRepository
    undefined, // runnerRepository
    undefined, // buildInfoRepository
    undefined, // sshAccessRepository
    stub.runnerService, // runnerService
    undefined, // volumeService
    undefined, // configService
    undefined, // warmPoolService
    undefined, // eventEmitter
    stub.organizationService, // organizationService
    stub.runnerAdapterFactory, // runnerAdapterFactory
    undefined, // organizationUsageService
    undefined, // redisLockProvider
    undefined, // redis
    stub.regionService, // regionService
    stub.snapshotService, // snapshotService
    undefined, // sandboxLookupCacheInvalidationService
    undefined, // sandboxActivityService
    stub.dockerRegistryService, // dockerRegistryService
    undefined, // sandboxForkRepository
    undefined, // sandboxSearchAdapter
  )

  return { service, stub }
}

describe('SandboxService.createSnapshotFromSandbox', () => {
  it.each([
    [
      'runner adapter creation throws',
      (stub: StubDeps) => {
        stub.runnerAdapterFactory.create.mockRejectedValue(new Error('adapter exploded'))
      },
      'adapter exploded',
    ],
    [
      'v2 snapshot job dispatch throws',
      (stub: StubDeps) => {
        stub.runnerAdapterFactory.create.mockResolvedValue({
          createSnapshotFromSandbox: jest.fn().mockRejectedValue(new Error('dispatch exploded')),
        })
      },
      'dispatch exploded',
    ],
  ])('restores sandbox state and fails the capture record when %s', async (_label, arrange, message) => {
    const { service, stub } = createService()
    jest.spyOn(service, 'findOneByIdOrName').mockResolvedValue(makeSandbox())
    arrange(stub)

    await expect(service.createSnapshotFromSandbox('sandbox-1', organization, { name: 'my-snap' })).rejects.toThrow(
      message,
    )

    // First call transitions to SNAPSHOTTING; second restores the previous state.
    expect(stub.sandboxRepository.updateWhere).toHaveBeenCalledTimes(2)
    expect(stub.sandboxRepository.updateWhere).toHaveBeenNthCalledWith(2, 'sandbox-1', {
      updateData: { state: SandboxState.STARTED, pending: false },
      whereCondition: { state: SandboxState.SNAPSHOTTING },
    })
    expect(stub.snapshotService.failSnapshotFromSandbox).toHaveBeenCalledWith({
      organizationId: 'org-1',
      name: 'my-snap',
      errorReason: message,
    })
    // The pending quota counter already transferred to the record lifecycle,
    // so the outer rollback must be a no-op.
    expect(stub.snapshotService.rollbackPendingUsage).toHaveBeenCalledWith('org-1', undefined)
  })

  it('returns the SNAPSHOTTING sandbox without touching the record when dispatch succeeds', async () => {
    const { service, stub } = createService()
    jest.spyOn(service, 'findOneByIdOrName').mockResolvedValue(makeSandbox())
    stub.runnerAdapterFactory.create.mockResolvedValue({
      createSnapshotFromSandbox: jest.fn(async () => undefined),
    })

    const result = await service.createSnapshotFromSandbox('sandbox-1', organization, { name: 'my-snap' })

    expect(result.state).toBe(SandboxState.SNAPSHOTTING)
    expect(stub.sandboxRepository.updateWhere).toHaveBeenCalledTimes(1)
    expect(stub.snapshotService.failSnapshotFromSandbox).not.toHaveBeenCalled()
  })
})
