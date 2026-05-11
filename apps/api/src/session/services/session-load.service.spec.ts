/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SessionLoadService, DaemonLoadSnapshot } from './session-load.service'
import { TypedConfigService } from '../../config/typed-config.service'

/** Minimal in-memory Redis covering the commands SessionLoadService uses. */
class FakeRedis {
  private store = new Map<string, string>()
  private sets = new Map<string, Set<string>>()
  private expiries = new Map<string, number>()

  async incr(key: string): Promise<number> {
    this.cleanExpired(key)
    const n = (parseInt(this.store.get(key) ?? '0', 10) || 0) + 1
    this.store.set(key, String(n))
    return n
  }
  async decr(key: string): Promise<number> {
    this.cleanExpired(key)
    const n = (parseInt(this.store.get(key) ?? '0', 10) || 0) - 1
    this.store.set(key, String(n))
    return n
  }
  async expire(key: string, seconds: number): Promise<number> {
    if (!this.store.has(key) && !this.sets.has(key)) return 0
    this.expiries.set(key, Date.now() + seconds * 1000)
    return 1
  }
  async get(key: string): Promise<string | null> {
    this.cleanExpired(key)
    return this.store.get(key) ?? null
  }
  async set(key: string, val: string): Promise<'OK'> {
    this.store.set(key, val)
    this.expiries.delete(key)
    return 'OK'
  }
  async sadd(key: string, member: string): Promise<number> {
    this.cleanExpired(key)
    const s = this.sets.get(key) ?? new Set<string>()
    const had = s.has(member)
    s.add(member)
    this.sets.set(key, s)
    return had ? 0 : 1
  }
  async srem(key: string, member: string): Promise<number> {
    this.cleanExpired(key)
    const s = this.sets.get(key)
    if (!s) return 0
    return s.delete(member) ? 1 : 0
  }

  private cleanExpired(key: string): void {
    const exp = this.expiries.get(key)
    if (exp !== undefined && Date.now() >= exp) {
      this.store.delete(key)
      this.sets.delete(key)
      this.expiries.delete(key)
    }
  }
}

const SCALE_DEFAULTS: Record<string, number> = {
  'session.scale.targetConcurrencyPerSandbox': 4,
  'session.scale.loadTtlSeconds': 30,
  'session.scale.loadPollMs': 5000,
  'session.scale.cpuPressureThreshold': 50,
  'session.scale.memUtilThreshold': 0.85,
  'session.scale.diskUtilThreshold': 0.9,
}

function makeConfig(overrides: Record<string, number> = {}): TypedConfigService {
  const map = { ...SCALE_DEFAULTS, ...overrides }
  return { get: (key: string) => map[key] } as unknown as TypedConfigService
}

function newService(config = makeConfig()): { svc: SessionLoadService; redis: FakeRedis } {
  const redis = new FakeRedis()
  const svc = new SessionLoadService(
    {} as any, // SessionInstanceStore (only used by the poller, not exercised here)
    redis as any,
    {} as any, // runnerService (poller only)
    config,
  )
  return { svc, redis }
}

describe('SessionLoadService', () => {
  describe('in-flight counters', () => {
    it('increments and decrements, never going negative', async () => {
      const { svc, redis } = newService()
      expect(await svc.incrInflight('i1')).toBe(1)
      expect(await svc.incrInflight('i1')).toBe(2)
      expect(await svc.getInflight('i1')).toBe(2)
      await svc.decrInflight('i1')
      expect(await svc.getInflight('i1')).toBe(1)
      await svc.decrInflight('i1')
      await svc.decrInflight('i1') // would go to -1
      expect(await svc.getInflight('i1')).toBe(0)
      const stored = await redis.get('session:load:inflight:i1')
      expect(stored).toBe('0')
      expect(await svc.incrInflight('i1')).toBe(1)
    })
  })

  describe('slot checkout', () => {
    it('hands out distinct slots until exhausted, then returns -1', async () => {
      const { svc } = newService()
      expect(await svc.checkoutSlot('i1', 'python', 3)).toBe(0)
      expect(await svc.checkoutSlot('i1', 'python', 3)).toBe(1)
      expect(await svc.checkoutSlot('i1', 'python', 3)).toBe(2)
      expect(await svc.checkoutSlot('i1', 'python', 3)).toBe(-1) // all taken -> caller uses ephemeral ctx
      await svc.releaseSlot('i1', 'python', 1)
      expect(await svc.checkoutSlot('i1', 'python', 3)).toBe(1) // freed slot reused
    })

    it('allocates no slot when maxSlots is 0 (forces ephemeral fallback)', async () => {
      const { svc } = newService()
      // [0, 0) -> no iterations -> -1 so the caller uses a unique ephemeral context.
      expect(await svc.checkoutSlot('i1', 'python', 0)).toBe(-1)
    })

    it('expires the slot set key so a long-lived slot is eventually reclaimed', async () => {
      jest.useFakeTimers()
      try {
        // loadTtl is small but execTimeout drives the slot TTL (>= execTimeout + 5s margin).
        const { svc } = newService(
          makeConfig({ 'session.scale.loadTtlSeconds': 30, 'session.execTimeoutSeconds': 600 }),
        )
        expect(await svc.checkoutSlot('i1', 'python', 1)).toBe(0)
        // Set key persists past the load TTL (would have been wrongly reclaimed before the fix).
        jest.setSystemTime(Date.now() + 60 * 1000)
        expect(await svc.checkoutSlot('i1', 'python', 1)).toBe(-1) // still in use
        // ...but does expire once past the exec timeout, freeing the slot for reuse.
        jest.setSystemTime(Date.now() + 605 * 1000)
        expect(await svc.checkoutSlot('i1', 'python', 1)).toBe(0)
      } finally {
        jest.useRealTimers()
      }
    })
  })

  describe('incrInflight fail-closed', () => {
    it('returns a negative sentinel when Redis is unavailable', async () => {
      const { svc, redis } = newService()
      jest.spyOn(redis, 'incr').mockRejectedValueOnce(new Error('redis down'))
      expect(await svc.incrInflight('i1')).toBeLessThan(0)
    })
  })

  describe('effectiveLoad', () => {
    it('is the max of in-flight and daemon busy contexts', async () => {
      const { svc, redis } = newService()
      await svc.incrInflight('i1') // inflight = 1
      // Distinct active vs busy so a field-swap (reading activeContexts instead of
      // busyContexts) would change the result and fail the test.
      const snap: DaemonLoadSnapshot = { activeContexts: 5, busyContexts: 3, pyMax: 16, tsMax: 64, bashMax: 16 }
      await redis.set('session:load:res:i1', JSON.stringify(snap))
      expect(await svc.effectiveLoad('i1')).toBe(3)
    })
  })

  describe('isResourceSaturated', () => {
    it('is false when no snapshot', () => {
      const { svc } = newService()
      expect(svc.isResourceSaturated(null)).toBe(false)
    })

    it('trips on CPU PSI pressure', () => {
      const { svc } = newService()
      const snap: DaemonLoadSnapshot = {
        activeContexts: 1,
        busyContexts: 0,
        pyMax: 16,
        tsMax: 64,
        bashMax: 16,
        cpu: { pressureSomeAvg10: 60 },
      }
      expect(svc.isResourceSaturated(snap)).toBe(true)
    })

    it('trips on memory utilization', () => {
      const { svc } = newService()
      const snap: DaemonLoadSnapshot = {
        activeContexts: 1,
        busyContexts: 0,
        pyMax: 16,
        tsMax: 64,
        bashMax: 16,
        memory: { utilization: 0.9 },
      }
      expect(svc.isResourceSaturated(snap)).toBe(true)
    })

    it('is false under all thresholds', () => {
      const { svc } = newService()
      const snap: DaemonLoadSnapshot = {
        activeContexts: 1,
        busyContexts: 0,
        pyMax: 16,
        tsMax: 64,
        bashMax: 16,
        cpu: { pressureSomeAvg10: 10 },
        memory: { utilization: 0.4 },
        disk: { utilization: 0.5 },
      }
      expect(svc.isResourceSaturated(snap)).toBe(false)
    })
  })

  describe('isSaturated', () => {
    it('trips when effective load reaches the concurrency target', async () => {
      const { svc } = newService(makeConfig({ 'session.scale.targetConcurrencyPerSandbox': 2 }))
      await svc.incrInflight('i1')
      await svc.incrInflight('i1')
      expect(await svc.isSaturated('i1')).toBe(true)
    })

    it('is false with headroom and no resource pressure', async () => {
      const { svc } = newService(makeConfig({ 'session.scale.targetConcurrencyPerSandbox': 4 }))
      await svc.incrInflight('i1')
      expect(await svc.isSaturated('i1')).toBe(false)
    })
  })
})
