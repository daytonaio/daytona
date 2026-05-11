/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SessionScheduler } from './session-scheduler.service'
import { SessionInstance } from '../entities/session-instance.entity'
import { SessionLoadService, DaemonLoadSnapshot } from './session-load.service'
import { TypedConfigService } from '../../config/typed-config.service'

function inst(id: string): SessionInstance {
  const e = new SessionInstance()
  e.id = id
  return e
}

function makeConfig(target = 4): TypedConfigService {
  return {
    get: (key: string) => (key === 'session.scale.targetConcurrencyPerSandbox' ? target : undefined),
  } as unknown as TypedConfigService
}

/**
 * Fake load service: `loads` is the pre-claim effective load per instance; incrInflight returns
 * the post-increment counter (seeded from loads) so the scheduler's atomic-claim logic can be
 * exercised deterministically.
 */
function makeLoad(loads: Record<string, number>, snaps: Record<string, DaemonLoadSnapshot> = {}) {
  const counters: Record<string, number> = { ...loads }
  return {
    effectiveLoad: jest.fn(async (id: string) => loads[id] ?? 0),
    getSnapshot: jest.fn(async (id: string) => snaps[id] ?? null),
    isResourceSaturated: jest.fn((snap: DaemonLoadSnapshot | null) => !!snap && (snap as any).__saturated === true),
    incrInflight: jest.fn(async (id: string) => {
      counters[id] = (counters[id] ?? 0) + 1
      return counters[id]
    }),
    decrInflight: jest.fn(async (id: string) => {
      counters[id] = (counters[id] ?? 0) - 1
    }),
    counters,
  } as unknown as SessionLoadService & { counters: Record<string, number> }
}

describe('SessionScheduler', () => {
  describe('claim', () => {
    it('returns null when there are no instances', async () => {
      const load = makeLoad({})
      const sched = new SessionScheduler(load, makeConfig())
      expect(await sched.claim([])).toBeNull()
    })

    it('claims the least-loaded instance with headroom', async () => {
      const load = makeLoad({ a: 3, b: 1 })
      const sched = new SessionScheduler(load, makeConfig(4))
      const picked = await sched.claim([inst('a'), inst('b')])
      expect(picked?.id).toBe('b')
      // b was incremented and kept; a was never touched.
      expect((load as any).counters.b).toBe(2)
      expect((load as any).counters.a).toBe(3)
    })

    it('returns null and releases claims when every instance is at the concurrency target', async () => {
      const load = makeLoad({ a: 4, b: 4 })
      const sched = new SessionScheduler(load, makeConfig(4))
      expect(await sched.claim([inst('a'), inst('b')])).toBeNull()
      // both were incr'd then decr'd back to their original loads.
      expect((load as any).counters.a).toBe(4)
      expect((load as any).counters.b).toBe(4)
    })

    it('treats daemon busy-context count as the floor of effective load', async () => {
      // Logical in-flight is 0 but the daemon reports 5 busy contexts (e.g. SDK-direct streams).
      const snaps: Record<string, DaemonLoadSnapshot> = {
        a: { activeContexts: 5, busyContexts: 5, pyMax: 16, tsMax: 64, bashMax: 16 },
      }
      const load = makeLoad({ a: 0 }, snaps)
      const sched = new SessionScheduler(load, makeConfig(4))
      expect(await sched.claim([inst('a')])).toBeNull()
    })

    it('skips instances under resource pressure even with concurrency headroom', async () => {
      const snaps: Record<string, DaemonLoadSnapshot> = {
        a: { activeContexts: 0, busyContexts: 0, pyMax: 16, tsMax: 64, __saturated: true } as any,
      }
      const load = makeLoad({ a: 0 }, snaps)
      const sched = new SessionScheduler(load, makeConfig(4))
      expect(await sched.claim([inst('a')])).toBeNull()
    })
  })

  describe('claimForce', () => {
    it('claims the least-loaded instance regardless of saturation', async () => {
      const load = makeLoad({ a: 6, b: 5 })
      const sched = new SessionScheduler(load, makeConfig(4))
      const picked = await sched.claimForce([inst('a'), inst('b')])
      expect(picked?.id).toBe('b')
      expect((load as any).counters.b).toBe(6)
    })

    it('returns null with no instances', async () => {
      const sched = new SessionScheduler(makeLoad({}), makeConfig())
      expect(await sched.claimForce([])).toBeNull()
    })
  })
})
