/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { getStateChangeLockKey } from '../utils/lock-key.util'
import { SandboxManager } from './sandbox.manager'

describe('SandboxManager automatic lifecycle checks', () => {
  const now = new Date('2026-09-03T02:14:40.000Z')

  function createManager(lastActivityAt: Date) {
    const sandbox = {
      id: 'sandbox-1',
      state: SandboxState.STARTED,
      autoStopInterval: 15,
      autoArchiveInterval: 15,
      autoDeleteInterval: 15,
    }
    const queryBuilder = {
      innerJoin: jest.fn().mockReturnThis(),
      where: jest.fn().mockReturnThis(),
      andWhere: jest.fn().mockReturnThis(),
      orderBy: jest.fn().mockReturnThis(),
      limit: jest.fn().mockReturnThis(),
      getMany: jest.fn().mockResolvedValue([sandbox]),
    }
    const sandboxRepository = {
      createQueryBuilder: jest.fn().mockReturnValue(queryBuilder),
      updateWhere: jest.fn().mockResolvedValue(sandbox),
    }
    const sandboxActivityService = {
      getLastActivityAt: jest.fn().mockResolvedValue(lastActivityAt),
    }
    const runnerService = {
      findAllReady: jest.fn().mockResolvedValue([{ id: 'runner-1' }]),
    }
    const redisLockProvider = {
      lock: jest.fn().mockResolvedValue(true),
      unlock: jest.fn().mockResolvedValue(undefined),
    }
    const configService = {
      get: jest.fn().mockReturnValue(undefined),
    }
    const manager = new SandboxManager(
      sandboxRepository as never,
      sandboxActivityService as never,
      runnerService as never,
      redisLockProvider as never,
      {} as never,
      {} as never,
      {} as never,
      {} as never,
      configService as never,
      {} as never,
      {} as never,
      {} as never,
      {} as never,
      {} as never,
      {} as never,
      {} as never,
    )
    jest.spyOn(manager, 'syncInstanceState').mockResolvedValue(undefined)

    return {
      manager,
      sandboxActivityService,
      sandboxRepository,
      redisLockProvider,
    }
  }

  beforeEach(() => {
    jest.spyOn(Date, 'now').mockReturnValue(now.getTime())
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  it('skips automatic lifecycle actions when buffered activity is still fresh', async () => {
    const { manager, sandboxActivityService, sandboxRepository, redisLockProvider } = createManager(
      new Date(now.getTime() - 60 * 1000),
    )

    await manager.autostopCheck()
    await manager.autoArchiveCheck()
    await manager.autoDeleteCheck()

    expect(sandboxActivityService.getLastActivityAt).toHaveBeenCalledTimes(3)
    expect(sandboxRepository.updateWhere).not.toHaveBeenCalled()
    expect(redisLockProvider.lock).toHaveBeenCalledWith(getStateChangeLockKey('sandbox-1'), 30)
    expect(sandboxActivityService.getLastActivityAt.mock.invocationCallOrder[0]).toBeGreaterThan(
      redisLockProvider.lock.mock.invocationCallOrder[1],
    )
  })

  it('stops a sandbox when the latest activity is expired', async () => {
    const { manager, sandboxRepository } = createManager(new Date(now.getTime() - 16 * 60 * 1000))

    await manager.autostopCheck()

    expect(sandboxRepository.updateWhere).toHaveBeenCalledWith('sandbox-1', {
      updateData: {
        pending: true,
        desiredState: SandboxDesiredState.STOPPED,
      },
      whereCondition: {
        pending: false,
        state: SandboxState.STARTED,
      },
    })
  })
})
