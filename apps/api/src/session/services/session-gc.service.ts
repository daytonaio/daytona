/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Cron, CronExpression } from '@nestjs/schedule'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { SessionState } from '../enums/session-state.enum'
import { TypedConfigService } from '../../config/typed-config.service'
import { sessionKeys } from './session-repository.service'

/** Minimal view of a stored context blob the GC needs to clean up secondary indexes. */
interface GcContext {
  orgId: string
  instanceId: string
  state: SessionState
}

/**
 * Atomically expire ONE due context. Re-checks the (possibly concurrently
 * extended) expiry score and the live state inside Redis before flipping
 * active → expired, so a context that touchLastUsed refreshed in the race window
 * between the candidate scan and the write is skipped rather than wrongly
 * expired. Returns 1 when it expired the context, else 0.
 *   KEYS[1]=ctx blob  KEYS[2]=gc expiry zset  KEYS[3]=gc grace zset
 *   ARGV: id, now, expiredAtIso, graceDeadline, activeState, expiredState, orgIndexPrefix
 */
export const SESSION_EXPIRE_SCRIPT = `
-- @daytona-session-expire
local score = redis.call('ZSCORE', KEYS[2], ARGV[1])
if not score then return 0 end
if tonumber(score) > tonumber(ARGV[2]) then return 0 end
local raw = redis.call('GET', KEYS[1])
if not raw then
  redis.call('ZREM', KEYS[2], ARGV[1])
  return 0
end
local ok, blob = pcall(cjson.decode, raw)
if not ok then
  redis.call('ZREM', KEYS[2], ARGV[1])
  return 0
end
if blob.state ~= ARGV[5] then
  redis.call('ZREM', KEYS[2], ARGV[1])
  return 0
end
blob.state = ARGV[6]
blob.expiredAt = ARGV[3]
redis.call('SET', KEYS[1], cjson.encode(blob))
redis.call('ZREM', KEYS[2], ARGV[1])
redis.call('ZADD', KEYS[3], ARGV[4], ARGV[1])
-- Guard the org-index key construction: concatenating a nil/non-string orgId
-- would raise mid-script AFTER the flip above already committed (Redis Lua does
-- not roll back), orphaning the index entry. A missing orgId just skips this
-- best-effort cleanup; list() self-heals a stale org-index member.
if type(blob.orgId) == 'string' then
  redis.call('ZREM', ARGV[7] .. blob.orgId, ARGV[1])
end
return 1
`

/**
 * SessionGcService enforces idle and absolute TTLs on Redis-backed contexts.
 *
 *  - sweepExpired() flips ACTIVE contexts whose expiry deadline has passed to EXPIRED. Candidates
 *    come from the `session:gc:expiry` zset (scored by the computed expiresAt, refreshed on every
 *    lastUsedAt touch), so a single ZRANGEBYSCORE up to `now` yields exactly the due contexts.
 *  - hardDeleteExpired() permanently removes EXPIRED/INVALID contexts past the grace period, taken
 *    from the `session:gc:grace` zset (scored by the grace deadline).
 *
 * Both crons run @EVERY_MINUTE; the grace zset preserves the 410 (expired/invalidated, with
 * reason) contract during the grace window. TTL knobs are re-read from process.env on every tick.
 */
@Injectable()
export class SessionGcService {
  private readonly logger = new Logger(SessionGcService.name)

  constructor(
    @InjectRedis()
    private readonly redis: Redis,
    private readonly config: TypedConfigService,
  ) {}

  @Cron(CronExpression.EVERY_MINUTE, { name: 'session-gc-sweep' })
  async sweepExpired(): Promise<void> {
    const batch = this.config.get('session.context.gcBatchSize') ?? 500
    const now = Date.now()
    const graceSec = this.intEnv(
      'SESSION_EXPIRED_GRACE_SECONDS',
      this.config.get('session.context.expiredGracePeriodSeconds') ?? 86400,
    )
    const graceDeadline = now + graceSec * 1000

    let ids: string[]
    try {
      ids = await this.redis.zrangebyscore(sessionKeys.gcExpiry, '-inf', now, 'LIMIT', 0, batch)
    } catch (err) {
      this.logger.error(`sweepExpired candidate scan failed: ${err.message}`)
      return
    }
    if (ids.length === 0) return

    // Flip each due context with an atomic CAS (SESSION_EXPIRE_SCRIPT) rather than
    // a read-then-write pipeline. The script re-validates the expiry score and the
    // live state inside Redis, which closes the TOCTOU with touchLastUsed: a
    // context refreshed after this candidate scan but before the flip is left
    // ACTIVE instead of being expired. Gone / already-non-ACTIVE contexts just get
    // their stale expiry-index entry dropped (the script handles both).
    const expiredAtIso = new Date(now).toISOString()
    const orgPrefix = sessionKeys.orgIndex('')
    const pipe = this.redis.pipeline()
    for (const id of ids) {
      pipe.eval(
        SESSION_EXPIRE_SCRIPT,
        3,
        sessionKeys.ctx(id),
        sessionKeys.gcExpiry,
        sessionKeys.gcGrace,
        id,
        String(now),
        expiredAtIso,
        String(graceDeadline),
        SessionState.ACTIVE,
        SessionState.EXPIRED,
        orgPrefix,
      )
    }

    let results: Array<[Error | null, unknown]> | null
    try {
      results = await pipe.exec()
    } catch (err) {
      this.logger.error(`sweepExpired pipeline failed: ${err.message}`)
      return
    }
    const settled = results ?? []
    const failed = settled.filter(([err]) => err).length
    if (failed > 0) {
      this.logger.warn(`sweepExpired: ${failed}/${ids.length} expire scripts errored (partial sweep)`)
    }
    const expired = settled.filter(([err, r]) => !err && Number(r) === 1).length
    if (expired > 0) this.logger.debug(`sweepExpired: marked ${expired} contexts EXPIRED`)
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'session-gc-hard-delete' })
  async hardDeleteExpired(): Promise<void> {
    const batch = this.config.get('session.context.gcBatchSize') ?? 500
    const now = Date.now()

    let ids: string[]
    try {
      ids = await this.redis.zrangebyscore(sessionKeys.gcGrace, '-inf', now, 'LIMIT', 0, batch)
    } catch (err) {
      this.logger.error(`hardDeleteExpired candidate scan failed: ${err.message}`)
      return
    }
    if (ids.length === 0) return

    let raws: (string | null)[]
    try {
      raws = await this.redis.mget(ids.map((id) => sessionKeys.ctx(id)))
    } catch (err) {
      this.logger.error(`hardDeleteExpired blob read failed: ${err.message}`)
      return
    }

    const pipe = this.redis.pipeline()
    for (let i = 0; i < ids.length; i++) {
      const id = ids[i]
      const raw = raws[i]
      pipe.del(sessionKeys.ctx(id))
      pipe.zrem(sessionKeys.gcGrace, id)
      if (raw) {
        const blob = JSON.parse(raw) as GcContext
        pipe.srem(sessionKeys.instanceContexts(blob.instanceId), id)
        pipe.zrem(sessionKeys.orgIndex(blob.orgId), id)
      }
    }

    try {
      await pipe.exec()
      this.logger.debug(`hardDeleteExpired: removed ${ids.length} contexts`)
    } catch (err) {
      this.logger.error(`hardDeleteExpired pipeline failed: ${err.message}`)
    }
  }

  private intEnv(name: string, fallback: number): number {
    const raw = process.env[name]
    if (!raw) return fallback
    const n = parseInt(raw, 10)
    return Number.isFinite(n) && n >= 0 ? n : fallback
  }
}
