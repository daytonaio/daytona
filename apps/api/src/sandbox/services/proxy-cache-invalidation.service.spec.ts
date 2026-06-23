/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import Redis from 'ioredis'

import { ProxyCacheInvalidationService } from './proxy-cache-invalidation.service'
import { SandboxPublicStatusUpdatedEvent } from '../events/sandbox-public-status-updated.event'
import { Sandbox } from '../entities/sandbox.entity'

describe('ProxyCacheInvalidationService', () => {
  let service: ProxyCacheInvalidationService
  let del: jest.Mock

  const SANDBOX_ID = 'sbx-1'
  const API_KEY = `preview:public:${SANDBOX_ID}`
  const PROXY_KEY = `proxy:sandbox-public:${SANDBOX_ID}`

  const makeEvent = () => new SandboxPublicStatusUpdatedEvent({ id: SANDBOX_ID } as Sandbox, true, false)

  beforeEach(() => {
    del = jest.fn().mockResolvedValue(1)
    service = new ProxyCacheInvalidationService({ del } as unknown as Redis)
  })

  describe('handleSandboxPublicStatusUpdated', () => {
    it('evicts both the API-side and the proxy-side public-status caches', async () => {
      await service.handleSandboxPublicStatusUpdated(makeEvent())

      expect(del).toHaveBeenCalledWith(API_KEY)
      expect(del).toHaveBeenCalledWith(PROXY_KEY)
    })

    // Ordering is the correctness property, not an incidental detail: the proxy only
    // re-queries the API on a cache miss, and a miss can only happen after the proxy
    // key is gone. If the proxy key were evicted first, a request landing in the gap
    // would re-read the still-cached API decision and re-populate the proxy's long-lived
    // cache. The API-side key must be evicted first.
    it('evicts the API-side cache before the proxy-side cache', async () => {
      await service.handleSandboxPublicStatusUpdated(makeEvent())

      const apiCallIndex = del.mock.calls.findIndex((args) => args[0] === API_KEY)
      const proxyCallIndex = del.mock.calls.findIndex((args) => args[0] === PROXY_KEY)

      expect(apiCallIndex).toBeGreaterThanOrEqual(0)
      expect(proxyCallIndex).toBeGreaterThanOrEqual(0)
      expect(del.mock.invocationCallOrder[apiCallIndex]).toBeLessThan(del.mock.invocationCallOrder[proxyCallIndex])
    })

    it('still evicts the proxy-side cache when the API-side eviction fails', async () => {
      del.mockImplementation((key: string) => {
        if (key === API_KEY) {
          return Promise.reject(new Error('redis down'))
        }
        return Promise.resolve(1)
      })

      await expect(service.handleSandboxPublicStatusUpdated(makeEvent())).resolves.not.toThrow()
      expect(del).toHaveBeenCalledWith(PROXY_KEY)
    })

    it('does not throw when both evictions fail (visibility change must not 500)', async () => {
      del.mockRejectedValue(new Error('redis down'))

      await expect(service.handleSandboxPublicStatusUpdated(makeEvent())).resolves.not.toThrow()
    })
  })
})
