/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SessionGcService, SESSION_EXPIRE_SCRIPT } from './session-gc.service'
import { sessionKeys, SESSION_TOUCH_SCRIPT } from './session-repository.service'
import { SessionState } from '../enums/session-state.enum'
import { TypedConfigService } from '../../config/typed-config.service'
import { FakeRedis } from './test-utils/fake-redis'

const ORG = 'org-1'
const INST = 'inst-1'

function makeConfig(overrides: Record<string, number> = {}): TypedConfigService {
  const map: Record<string, number> = {
    'session.context.gcBatchSize': 500,
    'session.context.expiredGracePeriodSeconds': 60,
    ...overrides,
  }
  return { get: (key: string) => map[key] } as unknown as TypedConfigService
}

function seedContext(redis: FakeRedis, id: string, state: SessionState, extra: Record<string, unknown> = {}): void {
  redis.strings.set(
    sessionKeys.ctx(id),
    JSON.stringify({
      id,
      orgId: ORG,
      instanceId: INST,
      language: 'python',
      state,
      createdAt: new Date().toISOString(),
      lastUsedAt: new Date().toISOString(),
      ...extra,
    }),
  )
}

describe('SessionGcService', () => {
  // FakeRedis.eval runs the REAL Lua scripts (via fengari), so these assert the
  // actual sweep/CAS behavior — including that the right args reach the right
  // positions (a swapped/missing arg would produce a wrong outcome and fail).
  describe('sweepExpired', () => {
    it('flips due ACTIVE contexts to EXPIRED and moves them to the grace zset', async () => {
      const redis = new FakeRedis()
      const gc = new SessionGcService(redis as any, makeConfig())
      const past = Date.now() - 1000

      seedContext(redis, 'c1', SessionState.ACTIVE)
      await redis.zadd(sessionKeys.gcExpiry, past, 'c1')
      await redis.zadd(sessionKeys.orgIndex(ORG), Date.now(), 'c1')

      await gc.sweepExpired()

      const blob = JSON.parse((await redis.get(sessionKeys.ctx('c1'))) ?? '{}')
      expect(blob.state).toBe(SessionState.EXPIRED)
      expect(blob.expiredAt).toBeDefined()
      // Dropped from active/expiry indexes, queued for hard delete.
      expect(await redis.zrangebyscore(sessionKeys.gcExpiry, '-inf', '+inf')).toHaveLength(0)
      expect(await redis.zrevrange(sessionKeys.orgIndex(ORG), 0, -1)).toHaveLength(0)
      expect(await redis.zrangebyscore(sessionKeys.gcGrace, '-inf', '+inf')).toEqual(['c1'])
    })

    it('leaves contexts whose expiry is still in the future untouched', async () => {
      const redis = new FakeRedis()
      const gc = new SessionGcService(redis as any, makeConfig())

      seedContext(redis, 'c1', SessionState.ACTIVE)
      await redis.zadd(sessionKeys.gcExpiry, Date.now() + 60_000, 'c1')

      await gc.sweepExpired()

      const blob = JSON.parse((await redis.get(sessionKeys.ctx('c1'))) ?? '{}')
      expect(blob.state).toBe(SessionState.ACTIVE)
      expect(await redis.zrangebyscore(sessionKeys.gcGrace, '-inf', '+inf')).toHaveLength(0)
    })
  })

  describe('hardDeleteExpired', () => {
    it('removes contexts past the grace deadline and cleans their indexes', async () => {
      const redis = new FakeRedis()
      const gc = new SessionGcService(redis as any, makeConfig())
      const past = Date.now() - 1000

      seedContext(redis, 'c1', SessionState.EXPIRED, { expiredAt: new Date(past).toISOString() })
      await redis.zadd(sessionKeys.gcGrace, past, 'c1')
      await redis.sadd(sessionKeys.instanceContexts(INST), 'c1')

      await gc.hardDeleteExpired()

      expect(await redis.get(sessionKeys.ctx('c1'))).toBeNull()
      expect(await redis.zrangebyscore(sessionKeys.gcGrace, '-inf', '+inf')).toHaveLength(0)
      expect(await redis.smembers(sessionKeys.instanceContexts(INST))).toHaveLength(0)
    })

    it('leaves contexts still within the grace window', async () => {
      const redis = new FakeRedis()
      const gc = new SessionGcService(redis as any, makeConfig())

      seedContext(redis, 'c1', SessionState.EXPIRED)
      await redis.zadd(sessionKeys.gcGrace, Date.now() + 60_000, 'c1')

      await gc.hardDeleteExpired()

      expect(await redis.get(sessionKeys.ctx('c1'))).not.toBeNull()
    })

    it('drops a grace entry whose blob is already gone without erroring', async () => {
      const redis = new FakeRedis()
      const gc = new SessionGcService(redis as any, makeConfig())
      // gcGrace points at an id whose ctx blob has disappeared (Redis-evicted or
      // externally deleted) — hard-delete must still clear the grace entry.
      await redis.zadd(sessionKeys.gcGrace, Date.now() - 1000, 'ghost')

      await gc.hardDeleteExpired()

      expect(await redis.zrangebyscore(sessionKeys.gcGrace, '-inf', '+inf')).toHaveLength(0)
    })
  })

  // These run the REAL Lua CAS scripts and deterministically construct the
  // interleaved state the sweep<->touch race produces, asserting the re-check
  // guards hold — the race semantics that have no other (e2e) coverage.
  describe('sweep/touch CAS race guards', () => {
    it('touch does NOT resurrect a context the GC already flipped to EXPIRED', async () => {
      const redis = new FakeRedis()
      seedContext(redis, 'c1', SessionState.EXPIRED)

      const r = await redis.eval(
        SESSION_TOUCH_SCRIPT,
        2,
        sessionKeys.ctx('c1'),
        sessionKeys.gcExpiry,
        'c1',
        new Date().toISOString(),
        String(Date.now() + 1000),
        SessionState.ACTIVE,
      )

      expect(r).toBe(0)
      expect(JSON.parse((await redis.get(sessionKeys.ctx('c1'))) ?? '{}').state).toBe(SessionState.EXPIRED)
    })

    it('touch refreshes lastUsedAt + expiry for an ACTIVE context', async () => {
      const redis = new FakeRedis()
      seedContext(redis, 'c1', SessionState.ACTIVE, { lastUsedAt: 'OLD' })

      const r = await redis.eval(
        SESSION_TOUCH_SCRIPT,
        2,
        sessionKeys.ctx('c1'),
        sessionKeys.gcExpiry,
        'c1',
        '2099-01-01T00:00:00.000Z',
        '99999',
        SessionState.ACTIVE,
      )

      expect(r).toBe(1)
      expect(JSON.parse((await redis.get(sessionKeys.ctx('c1'))) ?? '{}').lastUsedAt).toBe('2099-01-01T00:00:00.000Z')
      expect(await redis.zrangebyscore(sessionKeys.gcExpiry, '-inf', '+inf')).toEqual(['c1'])
    })

    it('expire SKIPS a context whose deadline a concurrent touch pushed into the future', async () => {
      const redis = new FakeRedis()
      const now = Date.now()
      seedContext(redis, 'c1', SessionState.ACTIVE)
      // The race: c1 was a due candidate, but a touch re-extended its expiry score
      // past `now` before the flip ran.
      await redis.zadd(sessionKeys.gcExpiry, now + 60_000, 'c1')

      const r = await redis.eval(
        SESSION_EXPIRE_SCRIPT,
        3,
        sessionKeys.ctx('c1'),
        sessionKeys.gcExpiry,
        sessionKeys.gcGrace,
        'c1',
        String(now),
        new Date(now).toISOString(),
        String(now + 1000),
        SessionState.ACTIVE,
        SessionState.EXPIRED,
        sessionKeys.orgIndex(''),
      )

      expect(r).toBe(0)
      expect(JSON.parse((await redis.get(sessionKeys.ctx('c1'))) ?? '{}').state).toBe(SessionState.ACTIVE)
      expect(await redis.zrangebyscore(sessionKeys.gcGrace, '-inf', '+inf')).toHaveLength(0)
    })

    it('expire drops a poisoned (undecodable) blob from the expiry zset instead of retrying forever', async () => {
      const redis = new FakeRedis()
      const now = Date.now()
      redis.strings.set(sessionKeys.ctx('c1'), '{not valid json')
      await redis.zadd(sessionKeys.gcExpiry, now - 1000, 'c1')

      const r = await redis.eval(
        SESSION_EXPIRE_SCRIPT,
        3,
        sessionKeys.ctx('c1'),
        sessionKeys.gcExpiry,
        sessionKeys.gcGrace,
        'c1',
        String(now),
        new Date(now).toISOString(),
        String(now + 1000),
        SessionState.ACTIVE,
        SessionState.EXPIRED,
        sessionKeys.orgIndex(''),
      )

      expect(r).toBe(0)
      // The poisoned entry is removed from gcExpiry so it isn't re-scanned every sweep.
      expect(await redis.zrangebyscore(sessionKeys.gcExpiry, '-inf', '+inf')).toHaveLength(0)
    })

    it('expire handles a blob with a missing orgId without erroring (skips org-index cleanup)', async () => {
      const redis = new FakeRedis()
      const now = Date.now()
      // No orgId — the org-index key would otherwise be `prefix .. nil`, which would
      // raise mid-script after the flip already committed.
      redis.strings.set(sessionKeys.ctx('c1'), JSON.stringify({ id: 'c1', state: SessionState.ACTIVE }))
      await redis.zadd(sessionKeys.gcExpiry, now - 1000, 'c1')

      const r = await redis.eval(
        SESSION_EXPIRE_SCRIPT,
        3,
        sessionKeys.ctx('c1'),
        sessionKeys.gcExpiry,
        sessionKeys.gcGrace,
        'c1',
        String(now),
        new Date(now).toISOString(),
        String(now + 1000),
        SessionState.ACTIVE,
        SessionState.EXPIRED,
        sessionKeys.orgIndex(''),
      )

      expect(r).toBe(1)
      expect(JSON.parse((await redis.get(sessionKeys.ctx('c1'))) ?? '{}').state).toBe(SessionState.EXPIRED)
      expect(await redis.zrangebyscore(sessionKeys.gcGrace, '-inf', '+inf')).toEqual(['c1'])
      expect(await redis.zrangebyscore(sessionKeys.gcExpiry, '-inf', '+inf')).toHaveLength(0)
    })
  })
})
