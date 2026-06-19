/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException, NotFoundException } from '@nestjs/common'
import { SandboxService } from './sandbox.service'
import { SandboxState } from '../enums/sandbox-state.enum'

describe('SandboxService.updateDegradedReason', () => {
  const reason = 'fd-exhaustion: too many open files'

  function makeService(sandbox: { id: string; state: SandboxState; degradedReason?: string | null } | null) {
    const sandboxRepository = {
      findOne: jest.fn().mockResolvedValue(sandbox),
      update: jest.fn().mockResolvedValue(sandbox),
    }
    const service = Object.create(SandboxService.prototype) as SandboxService
    Object.assign(service, { sandboxRepository })
    return { service, sandboxRepository }
  }

  it('persists the reason on a started sandbox', async () => {
    const sandbox = { id: 'sb', state: SandboxState.STARTED, degradedReason: null }
    const { service, sandboxRepository } = makeService(sandbox)

    await service.updateDegradedReason('sb', reason)

    expect(sandboxRepository.update).toHaveBeenCalledWith('sb', {
      updateData: { degradedReason: reason },
      entity: sandbox,
    })
  })

  it('throws 409 when setting a reason on a sandbox that is not started', async () => {
    const { service, sandboxRepository } = makeService({
      id: 'sb',
      state: SandboxState.STARTING,
      degradedReason: null,
    })

    await expect(service.updateDegradedReason('sb', reason)).rejects.toThrow(ConflictException)
    expect(sandboxRepository.update).not.toHaveBeenCalled()
  })

  it('accepts an idempotent clear on a sandbox that is not started', async () => {
    const { service, sandboxRepository } = makeService({
      id: 'sb',
      state: SandboxState.STOPPED,
      degradedReason: null,
    })

    await expect(service.updateDegradedReason('sb', null)).resolves.toBeUndefined()
    expect(sandboxRepository.update).not.toHaveBeenCalled()
  })

  it('persists a clear of a stale reason on a sandbox that is not started', async () => {
    const sandbox = { id: 'sb', state: SandboxState.STOPPED, degradedReason: reason }
    const { service, sandboxRepository } = makeService(sandbox)

    await expect(service.updateDegradedReason('sb', null)).resolves.toBeUndefined()
    expect(sandboxRepository.update).toHaveBeenCalledWith('sb', {
      updateData: { degradedReason: null },
      entity: sandbox,
    })
  })

  it('is a no-op when the reason is unchanged', async () => {
    const { service, sandboxRepository } = makeService({
      id: 'sb',
      state: SandboxState.STARTED,
      degradedReason: reason,
    })

    await expect(service.updateDegradedReason('sb', reason)).resolves.toBeUndefined()
    expect(sandboxRepository.update).not.toHaveBeenCalled()
  })

  it('throws 404 for an unknown sandbox', async () => {
    const { service } = makeService(null)

    await expect(service.updateDegradedReason('missing', reason)).rejects.toThrow(NotFoundException)
  })
})
