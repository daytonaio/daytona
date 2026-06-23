/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxEvents } from '../constants/sandbox-events.constants'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxAuthTokenRotatedEvent } from '../events/sandbox-auth-token-rotated.event'
import { SandboxRepository } from '../repositories/sandbox.repository'
import { ProxyCacheInvalidationService } from './proxy-cache-invalidation.service'

describe('[SANDBOX] preview auth token rotation cache invalidation', () => {
  describe('ProxyCacheInvalidationService', () => {
    function createService() {
      const redis = { del: jest.fn().mockResolvedValue(1) }
      const service = new ProxyCacheInvalidationService(redis as never)
      return { service, redis }
    }

    it('evicts both the API-side and proxy-side caches for the PREVIOUS token, not the new one', async () => {
      const { service, redis } = createService()

      await service.handleSandboxAuthTokenRotated(
        new SandboxAuthTokenRotatedEvent({ id: 'sandbox-1' } as Sandbox, 'old-token', 'new-token'),
      )

      // Must target the rotated-out token on BOTH cache layers so it stops authorizing immediately.
      // The API-side preview:token cache re-poisons the proxy cache on the next miss if left in place.
      expect(redis.del).toHaveBeenCalledWith('preview:token:sandbox-1:old-token')
      expect(redis.del).toHaveBeenCalledWith('proxy:sandbox-auth-key-valid:sandbox-1:old-token')
      // Deleting the new token would be a no-op and leave the stale entry alive.
      expect(redis.del).not.toHaveBeenCalledWith(expect.stringContaining('new-token'))
    })

    // Ordering is the correctness property: the proxy only re-queries the API on a cache miss,
    // and a miss can only happen after the proxy key is gone. If the proxy key were evicted first,
    // a request landing in the gap would re-validate against the still-cached API decision and
    // re-poison the proxy's longer-lived cache. The API-side key must be evicted first.
    it('evicts the API-side cache before the proxy-side cache', async () => {
      const { service, redis } = createService()

      await service.handleSandboxAuthTokenRotated(
        new SandboxAuthTokenRotatedEvent({ id: 'sandbox-1' } as Sandbox, 'old-token', 'new-token'),
      )

      const apiKey = 'preview:token:sandbox-1:old-token'
      const proxyKey = 'proxy:sandbox-auth-key-valid:sandbox-1:old-token'
      const apiCallIndex = redis.del.mock.calls.findIndex((args) => args[0] === apiKey)
      const proxyCallIndex = redis.del.mock.calls.findIndex((args) => args[0] === proxyKey)

      expect(apiCallIndex).toBeGreaterThanOrEqual(0)
      expect(proxyCallIndex).toBeGreaterThanOrEqual(0)
      expect(redis.del.mock.invocationCallOrder[apiCallIndex]).toBeLessThan(
        redis.del.mock.invocationCallOrder[proxyCallIndex],
      )
    })

    it('does nothing when there is no previous token', async () => {
      const { service, redis } = createService()

      await service.handleSandboxAuthTokenRotated(
        new SandboxAuthTokenRotatedEvent({ id: 'sandbox-1' } as Sandbox, '', 'new-token'),
      )

      expect(redis.del).not.toHaveBeenCalled()
    })

    it('does not throw if redis fails', async () => {
      const redis = { del: jest.fn().mockRejectedValue(new Error('redis down')) }
      const service = new ProxyCacheInvalidationService(redis as never)

      await expect(
        service.handleSandboxAuthTokenRotated(
          new SandboxAuthTokenRotatedEvent({ id: 'sandbox-1' } as Sandbox, 'old-token', 'new-token'),
        ),
      ).resolves.toBeUndefined()
    })
  })

  describe('SandboxRepository.emitUpdateEvents', () => {
    function createRepository() {
      const eventEmitter = { emit: jest.fn() }
      const dataSource = { getRepository: jest.fn().mockReturnValue({}) }
      const lookupCache = { invalidate: jest.fn(), invalidateOrgId: jest.fn() }
      const repository = new SandboxRepository(dataSource as never, eventEmitter as never, lookupCache as never)
      return { repository, eventEmitter }
    }

    const base = {
      id: 'sandbox-1',
      state: 'started',
      desiredState: 'started',
      public: false,
      organizationId: 'org-1',
    }

    function rotatedEvents(emit: jest.Mock) {
      return emit.mock.calls.filter((call) => call[0] === SandboxEvents.AUTH_TOKEN_ROTATED)
    }

    it('emits AUTH_TOKEN_ROTATED carrying the previous token when authToken changes', () => {
      const { repository, eventEmitter } = createRepository()

      const previous = { ...base, authToken: 'old-token' }
      const updated = { ...base, authToken: 'new-token' }

      ;(repository as unknown as { emitUpdateEvents: (u: unknown, p: unknown) => void }).emitUpdateEvents(
        updated,
        previous,
      )

      const calls = rotatedEvents(eventEmitter.emit)
      expect(calls).toHaveLength(1)
      const event = calls[0][1] as SandboxAuthTokenRotatedEvent
      expect(event.previousAuthToken).toBe('old-token')
      expect(event.newAuthToken).toBe('new-token')
    })

    it('does not emit AUTH_TOKEN_ROTATED when authToken is unchanged', () => {
      const { repository, eventEmitter } = createRepository()

      const previous = { ...base, authToken: 'same-token' }
      const updated = { ...base, authToken: 'same-token' }

      ;(repository as unknown as { emitUpdateEvents: (u: unknown, p: unknown) => void }).emitUpdateEvents(
        updated,
        previous,
      )

      expect(rotatedEvents(eventEmitter.emit)).toHaveLength(0)
    })
  })
})
