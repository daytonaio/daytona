/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { v4 as uuidv4 } from 'uuid'
import { Session } from '../entities/session.entity'
import { SessionInstance } from '../entities/session-instance.entity'
import { SessionState } from '../enums/session-state.enum'
import { SessionInstanceState } from '../enums/session-instance-state.enum'
import { SessionDto } from '../dto/session.dto'
import { SessionExpiredError, SessionInvalidatedError } from '../errors/session-errors'
import { TypedConfigService } from '../../config/typed-config.service'
import { SessionInstanceStore } from './session-instance-store.service'

interface StoredContext {
  id: string
  orgId: string
  instanceId: string
  language: string
  cwd?: string
  state: SessionState
  invalidatedAt?: string
  expiredAt?: string
  createdAt: string
  lastUsedAt: string
}

export interface ResolvedContext {
  context: Session
  instance: SessionInstance
}

/** Redis key scheme for Session contexts. Shared with SessionGcService. */
export const sessionKeys = {
  ctx: (id: string) => `session:ctx:${id}`,
  orgIndex: (orgId: string) => `session:org:${orgId}`,
  instanceContexts: (instanceId: string) => `session:inst:${instanceId}:ctxs`,
  gcExpiry: 'session:gc:expiry',
  gcGrace: 'session:gc:grace',
}

/**
 * Atomic "touch" CAS: refresh lastUsedAt + the GC expiry deadline ONLY if the
 * context is still active. The old read-modify-write rewrote the whole blob, which
 * could resurrect a context the GC sweep flipped to EXPIRED in the read→write
 * window; doing the state check and write atomically inside Redis prevents that.
 * Returns 1 if it refreshed, 0 otherwise.
 *   KEYS[1]=ctx blob  KEYS[2]=gc expiry zset
 *   ARGV: id, lastUsedAtIso, expiresAtScore, activeState
 */
export const SESSION_TOUCH_SCRIPT = `
-- @daytona-session-touch
local raw = redis.call('GET', KEYS[1])
if not raw then return 0 end
local ok, blob = pcall(cjson.decode, raw)
if not ok then return 0 end
if blob.state ~= ARGV[4] then return 0 end
blob.lastUsedAt = ARGV[2]
redis.call('SET', KEYS[1], cjson.encode(blob))
redis.call('ZADD', KEYS[2], ARGV[3], ARGV[1])
return 1
`

/**
 * SessionRepository is the **only** point that resolves a context-id to an in-sandbox target.
 *
 * Redis is the source of truth for Session contexts (no Postgres). The context blob lives at
 * `session:ctx:{id}`; secondary indexes drive listing, instance-scoped invalidation, and GC:
 *  - `session:org:{orgId}`            zset (score=createdAt) of ACTIVE context ids — for list().
 *  - `session:inst:{instanceId}:ctxs` set of context ids — for instance-roll invalidation.
 *  - `session:gc:expiry`              zset (score=expiresAt) of ACTIVE ids — GC sweep.
 *  - `session:gc:grace`              zset (score=grace deadline) of EXPIRED/INVALID ids — hard delete.
 *
 * Contexts are short-TTL by nature; the GC sweep flips ACTIVE→EXPIRED and the grace zset drives a
 * later hard delete, preserving the 410 (expired/invalidated, with reason) contract during grace.
 * `lastUsedAt` is bumped on a fire-and-forget basis, throttled to >= 5s per context.
 */
@Injectable()
export class SessionRepository {
  private readonly logger = new Logger(SessionRepository.name)
  private readonly inFlight = new Map<string, Promise<ResolvedContext>>()
  private readonly lastUsedTouch = new Map<string, number>()

  constructor(
    @InjectRedis()
    private readonly redis: Redis,
    private readonly config: TypedConfigService,
    private readonly instances: SessionInstanceStore,
  ) {}

  /** Computes expiresAt = min(lastUsedAt + idleTtl, createdAt + absoluteTtl). */
  computeExpiresAt(createdAt: Date | string, lastUsedAt: Date | string): Date {
    const created = typeof createdAt === 'string' ? new Date(createdAt) : createdAt
    const lastUsed = typeof lastUsedAt === 'string' ? new Date(lastUsedAt) : lastUsedAt
    // Read TTLs from process.env at every call so e2e tests can flip them without restart.
    const idleTtlSec = this.readIntEnv(
      'SESSION_IDLE_TTL_SECONDS',
      this.config.get('session.context.idleTtlSeconds') ?? 3600,
    )
    const absTtlSec = this.readIntEnv(
      'SESSION_ABSOLUTE_TTL_SECONDS',
      this.config.get('session.context.absoluteTtlSeconds') ?? 604800,
    )
    const idleExp = new Date(lastUsed.getTime() + idleTtlSec * 1000)
    const absExp = new Date(created.getTime() + absTtlSec * 1000)
    return idleExp < absExp ? idleExp : absExp
  }

  toDto(c: Session): SessionDto {
    return {
      id: c.id,
      language: c.language,
      cwd: c.cwd,
      createdAt: c.createdAt.toISOString(),
      lastUsedAt: c.lastUsedAt?.toISOString(),
      expiresAt: this.computeExpiresAt(c.createdAt, c.lastUsedAt ?? c.createdAt).toISOString(),
    }
  }

  async create(orgId: string, instance: SessionInstance, opts: { language: string; cwd?: string }): Promise<Session> {
    const id = uuidv4()
    const now = new Date()
    const ctx = new Session()
    ctx.id = id
    ctx.organizationId = orgId
    ctx.instanceId = instance.id
    ctx.language = opts.language
    ctx.cwd = opts.cwd
    ctx.state = SessionState.ACTIVE
    ctx.createdAt = now
    ctx.lastUsedAt = now

    const expiresAt = this.computeExpiresAt(now, now).getTime()
    await this.redis
      .pipeline()
      .set(sessionKeys.ctx(id), this.serialize(ctx))
      .zadd(sessionKeys.orgIndex(orgId), now.getTime(), id)
      .sadd(sessionKeys.instanceContexts(instance.id), id)
      .zadd(sessionKeys.gcExpiry, expiresAt, id)
      .exec()
    return ctx
  }

  async resolve(orgId: string, sessionId: string): Promise<ResolvedContext> {
    const inFlightKey = `${orgId}:${sessionId}`
    const existing = this.inFlight.get(inFlightKey)
    if (existing) return existing

    const promise = this.resolveInner(orgId, sessionId).finally(() => {
      this.inFlight.delete(inFlightKey)
    })
    this.inFlight.set(inFlightKey, promise)
    return promise
  }

  private async resolveInner(orgId: string, sessionId: string): Promise<ResolvedContext> {
    const ctx = await this.readContext(sessionId)
    if (!ctx) {
      throw new NotFoundException(`Session ${sessionId} not found.`)
    }
    this.assertOrgOwnership(ctx, orgId, sessionId)
    this.assertContextActive(ctx, sessionId)

    const inst = await this.instances.findById(ctx.instanceId)
    if (!inst || inst.state !== SessionInstanceState.READY) {
      throw new SessionInvalidatedError(sessionId, inst?.updatedAt ?? new Date())
    }

    const ctxEntity = this.toEntity(ctx)
    this.touchLastUsed(sessionId).catch((err) => this.logger.debug(`touchLastUsed: ${err.message}`))
    return { context: ctxEntity, instance: inst }
  }

  async delete(orgId: string, sessionId: string): Promise<void> {
    const ctx = await this.readContext(sessionId)
    if (!ctx || ctx.orgId !== orgId) return // idempotent / org-scoped
    try {
      await this.redis
        .pipeline()
        .del(sessionKeys.ctx(sessionId))
        .zrem(sessionKeys.orgIndex(orgId), sessionId)
        .srem(sessionKeys.instanceContexts(ctx.instanceId), sessionId)
        .zrem(sessionKeys.gcExpiry, sessionId)
        .zrem(sessionKeys.gcGrace, sessionId)
        .exec()
    } catch (err) {
      this.logger.warn(`redis cleanup on delete failed: ${err.message}`)
    }
  }

  async list(orgId: string, templateId?: string): Promise<SessionDto[]> {
    // Newest first (mirrors the previous ORDER BY createdAt DESC).
    const ids = await this.redis.zrevrange(sessionKeys.orgIndex(orgId), 0, -1)
    if (ids.length === 0) return []
    const raws = await this.redis.mget(ids.map((id) => sessionKeys.ctx(id)))

    // Collect the live ACTIVE contexts first; prune any dangling orgIndex member
    // whose blob has disappeared (e.g. Redis-evicted or externally deleted) so the
    // index self-heals rather than growing unbounded.
    const active: StoredContext[] = []
    const dangling: string[] = []
    for (let i = 0; i < raws.length; i++) {
      const raw = raws[i]
      if (!raw) {
        dangling.push(ids[i])
        continue
      }
      const ctx = JSON.parse(raw) as StoredContext
      if (ctx.state !== SessionState.ACTIVE) continue
      active.push(ctx)
    }
    if (dangling.length > 0) {
      this.redis
        .zrem(sessionKeys.orgIndex(orgId), ...dangling)
        .catch((err) => this.logger.debug(`list prune dangling orgIndex ids failed: ${err.message}`))
    }

    // When filtering by template, resolve every referenced instance in ONE batch
    // instead of a per-context findById (avoids an N+1 under large org listings).
    let instancesById: Map<string, SessionInstance> | undefined
    if (templateId) {
      const instanceIds = [...new Set(active.map((c) => c.instanceId))]
      instancesById = await this.instances.findByIds(instanceIds)
    }

    const out: SessionDto[] = []
    for (const ctx of active) {
      if (templateId) {
        const inst = instancesById?.get(ctx.instanceId)
        if (!inst || inst.templateId !== templateId) continue
      }
      out.push(this.toDto(this.toEntity(ctx)))
    }
    return out
  }

  /**
   * Bulk-marks all ACTIVE contexts for an instance as INVALID. Called by the pool reconciler when
   * a sandbox dies or a snapshot drift is detected.
   */
  async markInstanceSessionsInvalid(instanceId: string): Promise<void> {
    let ids: string[]
    try {
      ids = await this.redis.smembers(sessionKeys.instanceContexts(instanceId))
    } catch (err) {
      this.logger.warn(`markInstanceSessionsInvalid: reading instance contexts failed: ${err.message}`)
      return
    }
    if (ids.length === 0) return

    const now = new Date()
    const graceSec = this.readIntEnv(
      'SESSION_EXPIRED_GRACE_SECONDS',
      this.config.get('session.context.expiredGracePeriodSeconds') ?? 86400,
    )
    const graceDeadline = now.getTime() + graceSec * 1000

    const raws = await this.redis.mget(ids.map((id) => sessionKeys.ctx(id)))
    const pipe = this.redis.pipeline()
    for (let i = 0; i < raws.length; i++) {
      const raw = raws[i]
      const id = ids[i]
      if (!raw) {
        pipe.srem(sessionKeys.instanceContexts(instanceId), id)
        continue
      }
      const ctx = JSON.parse(raw) as StoredContext
      if (ctx.state !== SessionState.ACTIVE) continue
      ctx.state = SessionState.INVALID
      ctx.invalidatedAt = now.toISOString()
      pipe.set(sessionKeys.ctx(id), JSON.stringify(ctx))
      pipe.zrem(sessionKeys.orgIndex(ctx.orgId), id)
      pipe.zrem(sessionKeys.gcExpiry, id)
      pipe.zadd(sessionKeys.gcGrace, graceDeadline, id)
    }
    try {
      await pipe.exec()
    } catch (err) {
      this.logger.warn(`markInstanceSessionsInvalid: pipeline failed: ${err.message}`)
    }
  }

  // -- internals ------------------------------------------------------------

  private async readContext(sessionId: string): Promise<StoredContext | null> {
    const raw = await this.redis.get(sessionKeys.ctx(sessionId))
    if (!raw) return null
    return JSON.parse(raw) as StoredContext
  }

  private serialize(ctx: Session): string {
    const blob: StoredContext = {
      id: ctx.id,
      orgId: ctx.organizationId,
      instanceId: ctx.instanceId,
      language: ctx.language,
      cwd: ctx.cwd,
      state: ctx.state,
      invalidatedAt: ctx.invalidatedAt?.toISOString(),
      expiredAt: ctx.expiredAt?.toISOString(),
      createdAt: ctx.createdAt.toISOString(),
      lastUsedAt: ctx.lastUsedAt.toISOString(),
    }
    return JSON.stringify(blob)
  }

  private toEntity(c: StoredContext): Session {
    const e = new Session()
    e.id = c.id
    e.organizationId = c.orgId
    e.instanceId = c.instanceId
    e.language = c.language
    e.cwd = c.cwd
    e.state = c.state
    e.invalidatedAt = c.invalidatedAt ? new Date(c.invalidatedAt) : undefined
    e.expiredAt = c.expiredAt ? new Date(c.expiredAt) : undefined
    e.createdAt = new Date(c.createdAt)
    e.lastUsedAt = new Date(c.lastUsedAt)
    return e
  }

  private async touchLastUsed(sessionId: string): Promise<void> {
    const throttleMs = this.config.get('session.cache.lastUsedAtThrottleMs') ?? 5000
    const last = this.lastUsedTouch.get(sessionId) ?? 0
    const now = Date.now()
    if (now - last < throttleMs) return
    this.lastUsedTouch.set(sessionId, now)
    try {
      // Pre-read only to derive expiresAt from the (immutable) createdAt; the
      // mutation itself is an atomic CAS that re-checks state==ACTIVE inside Redis,
      // so it can't resurrect a context the GC sweep just flipped to EXPIRED.
      const ctx = await this.readContext(sessionId)
      if (!ctx || ctx.state !== SessionState.ACTIVE) return
      const nowDate = new Date(now)
      const expiresAt = this.computeExpiresAt(ctx.createdAt, nowDate).getTime()
      await this.redis.eval(
        SESSION_TOUCH_SCRIPT,
        2,
        sessionKeys.ctx(sessionId),
        sessionKeys.gcExpiry,
        sessionId,
        nowDate.toISOString(),
        String(expiresAt),
        SessionState.ACTIVE,
      )
    } catch (err) {
      this.logger.debug(`touchLastUsed: ${err.message}`)
    }
  }

  private assertOrgOwnership(c: StoredContext, orgId: string, sessionId: string): void {
    if (c.orgId !== orgId) {
      // Don't leak existence: surface 404, not 403.
      throw new NotFoundException(`Session ${sessionId} not found.`)
    }
  }

  private assertContextActive(c: StoredContext, sessionId: string): void {
    if (c.state === SessionState.INVALID) {
      throw new SessionInvalidatedError(sessionId, c.invalidatedAt ?? new Date().toISOString())
    }
    if (c.state === SessionState.EXPIRED) {
      const reason = this.classifyExpiry(c)
      throw new SessionExpiredError(sessionId, c.expiredAt ?? new Date().toISOString(), reason)
    }
  }

  private classifyExpiry(c: StoredContext): 'idle' | 'absolute' {
    const created = new Date(c.createdAt).getTime()
    const used = new Date(c.lastUsedAt).getTime()
    const idleTtl = this.readIntEnv(
      'SESSION_IDLE_TTL_SECONDS',
      this.config.get('session.context.idleTtlSeconds') ?? 3600,
    )
    const absTtl = this.readIntEnv(
      'SESSION_ABSOLUTE_TTL_SECONDS',
      this.config.get('session.context.absoluteTtlSeconds') ?? 604800,
    )
    return created + absTtl * 1000 <= used + idleTtl * 1000 ? 'absolute' : 'idle'
  }

  private readIntEnv(name: string, fallback: number): number {
    const raw = process.env[name]
    if (!raw) return fallback
    const n = parseInt(raw, 10)
    return Number.isFinite(n) ? n : fallback
  }
}
