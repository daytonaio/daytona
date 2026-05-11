/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnModuleDestroy, OnModuleInit } from '@nestjs/common'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { SessionInstance } from '../entities/session-instance.entity'
import { SessionInstanceState } from '../enums/session-instance-state.enum'
import { SessionInstanceStore } from './session-instance-store.service'
import { RunnerService } from '../../sandbox/services/runner.service'
import { TypedConfigService } from '../../config/typed-config.service'
import { buildDaemonAccess } from '../common/daemon-access'

/**
 * Snapshot of an in-sandbox daemon's self-reported load, as returned by `GET /load`.
 * Resource fields are optional: when the daemon can't read cgroup data it omits them and the
 * API falls back to concurrency-only saturation.
 */
export interface DaemonLoadSnapshot {
  activeContexts: number
  busyContexts: number
  pyMax: number
  tsMax: number
  // Optional: a warm sandbox running an older session-runtime image may return a
  // /load response that predates bashMax (rolling-deploy skew). The snapshot is an
  // unchecked cast of the daemon JSON, so the type must not claim a field that can
  // be absent at runtime.
  bashMax?: number
  cpu?: { utilization?: number; pressureSomeAvg10?: number }
  memory?: { utilization?: number; pressureSomeAvg10?: number }
  io?: { pressureSomeAvg10?: number }
  disk?: { utilization?: number }
}

const KEY_INFLIGHT = (id: string) => `session:load:inflight:${id}`
const KEY_RES = (id: string) => `session:load:res:${id}`

/**
 * SessionLoadService tracks how loaded each warm sandbox is, combining two signals:
 *
 *  1. An optimistic in-flight counter in Redis (`incr`/`decr` around each API-driven exec). It
 *     bridges the gap between load polls so concurrent picks across API replicas don't all pile
 *     onto the same "least-loaded" sandbox before the next poll observes the spike.
 *  2. A periodically polled snapshot of each daemon's `GET /load` — the source of truth for
 *     concurrency (it also counts SDK-direct `connect` streams the API never sees) plus
 *     cgroup-aware CPU/mem/disk pressure.
 *
 * `effectiveLoad = max(daemonBusyContexts, redisInflight)`. An instance is saturated when its
 * effective load reaches the per-sandbox target OR any resource-pressure threshold is crossed.
 */
@Injectable()
export class SessionLoadService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(SessionLoadService.name)
  private pollTimer?: ReturnType<typeof setInterval>

  constructor(
    private readonly instances: SessionInstanceStore,
    @InjectRedis()
    private readonly redis: Redis,
    private readonly runnerService: RunnerService,
    private readonly config: TypedConfigService,
  ) {}

  onModuleInit(): void {
    const pollMs = this.config.get('session.scale.loadPollMs') ?? 5000
    // Manual interval (not @Interval) so the period is config-driven.
    this.pollTimer = setInterval(() => {
      this.pollAll().catch((err) => this.logger.warn(`load poll failed: ${err.message}`))
    }, pollMs)
    // Don't keep the event loop alive solely for polling (matters for tests / graceful shutdown).
    this.pollTimer.unref?.()
  }

  onModuleDestroy(): void {
    if (this.pollTimer) clearInterval(this.pollTimer)
  }

  // -- in-flight counters --------------------------------------------------

  /**
   * Optimistically mark one more in-flight op on an instance; returns the new (monotonic) count.
   * Returns -1 (a sentinel below any valid post-incr count, which is always >= 1) when Redis is
   * unavailable, so the scheduler fails *closed* — a failed increment must never be interpreted as
   * a successful admission (effective load <= target).
   */
  async incrInflight(instanceId: string): Promise<number> {
    const ttl = this.config.get('session.scale.loadTtlSeconds') ?? 30
    try {
      const n = await this.redis.incr(KEY_INFLIGHT(instanceId))
      await this.redis.expire(KEY_INFLIGHT(instanceId), ttl)
      return n
    } catch (err) {
      this.logger.debug(`incrInflight(${instanceId}) failed: ${err.message}`)
      return -1
    }
  }

  /** Release one in-flight op on an instance. Never lets the counter go negative. */
  async decrInflight(instanceId: string): Promise<void> {
    try {
      const n = await this.redis.decr(KEY_INFLIGHT(instanceId))
      if (n < 0) await this.redis.set(KEY_INFLIGHT(instanceId), '0')
    } catch (err) {
      this.logger.debug(`decrInflight(${instanceId}) failed: ${err.message}`)
    }
  }

  // -- transient context slots --------------------------------------------

  /**
   * Check out a free transient-context slot in `[0, maxSlots)` for (instance, language) so
   * concurrent one-shot ops run on distinct daemon contexts. Uses a Redis set of in-use slots.
   * Returns `-1` when every slot is taken; callers MUST then fall back to a unique ephemeral
   * context (never a shared slot — a daemon context only accepts one WS client at a time, so
   * sharing one concurrently evicts the other op's stream and yields empty output).
   */
  async checkoutSlot(instanceId: string, language: string, maxSlots: number): Promise<number> {
    const key = `session:slots:${instanceId}:${language}`
    // The slot key must outlive the longest possible exec — a slot that expires mid-exec could be
    // re-handed-out and collide two ops on one daemon context. Use the larger of the load TTL and
    // the max exec timeout (plus the same +5s margin runOnDaemon applies to its socket deadline).
    const loadTtl = this.config.get('session.scale.loadTtlSeconds') ?? 30
    const execTtl = (this.config.get('session.execTimeoutSeconds') ?? 600) + 5
    const ttl = Math.max(loadTtl, execTtl)
    try {
      // [0, maxSlots): when maxSlots === 0 no slot is allocated (caller uses an ephemeral context).
      for (let slot = 0; slot < maxSlots; slot++) {
        const added = await this.redis.sadd(key, String(slot))
        if (added === 1) {
          await this.redis.expire(key, ttl)
          return slot
        }
      }
    } catch (err) {
      this.logger.debug(`checkoutSlot(${instanceId}) failed: ${err.message}`)
    }
    return -1
  }

  async releaseSlot(instanceId: string, language: string, slot: number): Promise<void> {
    try {
      await this.redis.srem(`session:slots:${instanceId}:${language}`, String(slot))
    } catch (err) {
      this.logger.debug(`releaseSlot(${instanceId}) failed: ${err.message}`)
    }
  }

  async getInflight(instanceId: string): Promise<number> {
    try {
      const raw = await this.redis.get(KEY_INFLIGHT(instanceId))
      const n = raw ? parseInt(raw, 10) : 0
      return Number.isFinite(n) && n > 0 ? n : 0
    } catch {
      return 0
    }
  }

  async getSnapshot(instanceId: string): Promise<DaemonLoadSnapshot | null> {
    try {
      const raw = await this.redis.get(KEY_RES(instanceId))
      return raw ? (JSON.parse(raw) as DaemonLoadSnapshot) : null
    } catch {
      return null
    }
  }

  // -- saturation / ordering ----------------------------------------------

  /**
   * Effective load = max(daemon-reported busy contexts, optimistic in-flight). Used both to
   * pick the least-loaded instance and to decide saturation.
   */
  async effectiveLoad(instanceId: string): Promise<number> {
    const [inflight, snap] = await Promise.all([this.getInflight(instanceId), this.getSnapshot(instanceId)])
    return Math.max(inflight, snap?.busyContexts ?? 0)
  }

  /** True when an instance should be considered full (concurrency target or resource pressure). */
  async isSaturated(instanceId: string): Promise<boolean> {
    const target = this.config.get('session.scale.targetConcurrencyPerSandbox') ?? 4
    const load = await this.effectiveLoad(instanceId)
    if (load >= target) return true
    return this.isResourceSaturated(await this.getSnapshot(instanceId))
  }

  isResourceSaturated(snap: DaemonLoadSnapshot | null): boolean {
    if (!snap) return false
    const cpuPressureThreshold = this.config.get('session.scale.cpuPressureThreshold') ?? 50
    const memUtilThreshold = this.config.get('session.scale.memUtilThreshold') ?? 0.85
    const diskUtilThreshold = this.config.get('session.scale.diskUtilThreshold') ?? 0.9

    if ((snap.cpu?.pressureSomeAvg10 ?? 0) >= cpuPressureThreshold) return true
    if ((snap.memory?.utilization ?? 0) >= memUtilThreshold) return true
    if ((snap.disk?.utilization ?? 0) >= diskUtilThreshold) return true
    return false
  }

  // -- polling -------------------------------------------------------------

  private async pollAll(): Promise<void> {
    const instances = await this.instances.findByState(SessionInstanceState.READY)
    if (instances.length === 0) return
    await Promise.all(instances.map((inst) => this.pollOne(inst).catch(() => undefined)))
  }

  private async pollOne(inst: SessionInstance): Promise<void> {
    if (!inst.sandboxId) return
    const port = this.config.get('session.daemonPort') ?? 2281
    const ttl = this.config.get('session.scale.loadTtlSeconds') ?? 30
    let runner: Awaited<ReturnType<RunnerService['findBySandboxId']>>
    try {
      runner = await this.runnerService.findBySandboxId(inst.sandboxId)
    } catch {
      return // sandbox gone / not found — reconcile will roll it
    }
    let access
    try {
      access = buildDaemonAccess(runner, inst.sandboxId, port)
    } catch {
      return
    }

    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), 3000)
    try {
      const resp = await fetch(`${access.url}/load`, {
        headers: { Authorization: `Bearer ${access.runnerApiKey}` },
        signal: controller.signal,
      })
      if (!resp.ok) return
      const snap = (await resp.json()) as DaemonLoadSnapshot
      await this.redis.set(KEY_RES(inst.id), JSON.stringify(snap), 'EX', ttl)
    } catch (err) {
      this.logger.debug(`load poll for instance ${inst.id} failed: ${(err as Error).message}`)
    } finally {
      clearTimeout(timer)
    }
  }
}
