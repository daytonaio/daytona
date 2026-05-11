/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { v4 as uuidv4 } from 'uuid'
import { SessionInstance } from '../entities/session-instance.entity'
import { SessionInstanceState } from '../enums/session-instance-state.enum'
import { SessionInstanceRole } from '../enums/session-instance-role.enum'

// Cap how many keys a single MGET batches, so a large id list can't issue one
// unbounded command / response (per-command payload + memory pressure).
// findByIds loops over chunks of this size.
const MGET_CHUNK = 256

/**
 * Redis persistence for SessionInstance — the warm-sandbox fleet's pool state. Replaces the
 * former Postgres table.
 *
 * Data model:
 *  - `session:inst:{id}`                                  JSON blob (source of truth).
 *  - `session:inst:state:{state}`                         SET of instance ids (global by state).
 *  - `session:inst:org:{orgId}:tpl:{templateId}:state:{s}`SET of instance ids (per-fleet by state).
 *
 * Every mutation re-derives index membership from the previous blob, so the state sets stay
 * consistent. There is intentionally NO TTL on instance keys: a warm instance must persist for as
 * long as its sandbox is alive. Redis may still be wiped (treated as ephemeral) — the pool's
 * orphan-sandbox reconciler is what prevents a wipe from leaking the underlying sandboxes.
 */
@Injectable()
export class SessionInstanceStore {
  private readonly logger = new Logger(SessionInstanceStore.name)

  constructor(
    @InjectRedis()
    private readonly redis: Redis,
  ) {}

  private static KEY(id: string): string {
    return `session:inst:${id}`
  }

  private static STATE_IDX(state: SessionInstanceState): string {
    return `session:inst:state:${state}`
  }

  private static ORG_TPL_STATE_IDX(orgId: string, templateId: string, state: SessionInstanceState): string {
    return `session:inst:org:${orgId}:tpl:${templateId}:state:${state}`
  }

  async create(input: {
    organizationId: string
    templateId: string
    snapshotId?: string
    sandboxId?: string
    state?: SessionInstanceState
    role?: SessionInstanceRole
  }): Promise<SessionInstance> {
    const now = new Date()
    const inst = new SessionInstance()
    inst.id = uuidv4()
    inst.organizationId = input.organizationId
    inst.templateId = input.templateId
    inst.snapshotId = input.snapshotId as string
    inst.sandboxId = input.sandboxId
    inst.state = input.state ?? SessionInstanceState.PROVISIONING
    inst.role = input.role ?? SessionInstanceRole.WARM
    inst.createdAt = now
    inst.updatedAt = now
    await this.write(inst, null)
    return inst
  }

  /** Upsert an instance, re-indexing state membership against the persisted previous blob. */
  async save(inst: SessionInstance): Promise<SessionInstance> {
    const prev = await this.findById(inst.id)
    inst.updatedAt = new Date()
    await this.write(inst, prev)
    return inst
  }

  /** Best-effort partial update (e.g. lastActiveAt) that leaves state/index membership intact. */
  async update(id: string, patch: Partial<SessionInstance>): Promise<void> {
    const prev = await this.findById(id)
    if (!prev) return
    Object.assign(prev, patch)
    prev.updatedAt = new Date()
    await this.write(prev, prev)
  }

  async findById(id: string): Promise<SessionInstance | null> {
    const raw = await this.redis.get(SessionInstanceStore.KEY(id))
    return raw ? this.deserialize(raw) : null
  }

  /** Batch-load instances by id, returning a map keyed by id (missing blobs are
   *  omitted). MGET is chunked (MGET_CHUNK) so a large id list can't issue one
   *  unbounded command / response. Lets callers avoid an N+1 of findById. */
  async findByIds(ids: string[]): Promise<Map<string, SessionInstance>> {
    const out = new Map<string, SessionInstance>()
    for (let i = 0; i < ids.length; i += MGET_CHUNK) {
      const chunk = ids.slice(i, i + MGET_CHUNK)
      const raws = await this.redis.mget(chunk.map((id) => SessionInstanceStore.KEY(id)))
      raws.forEach((raw, j) => {
        if (raw) out.set(chunk[j], this.deserialize(raw))
      })
    }
    return out
  }

  async findByOrgTemplateState(
    orgId: string,
    templateId: string,
    state: SessionInstanceState,
  ): Promise<SessionInstance[]> {
    const indexKey = SessionInstanceStore.ORG_TPL_STATE_IDX(orgId, templateId, state)
    const ids = await this.redis.smembers(indexKey)
    // The blob is authoritative: filter to instances that actually still match
    // this (org, template, state) so a stale set entry — e.g. one a concurrent
    // save() left behind — never inflates the result or the cap count.
    return this.readIndex(
      indexKey,
      ids,
      (inst) => inst.state === state && inst.organizationId === orgId && inst.templateId === templateId,
    )
  }

  async countByState(orgId: string, templateId: string, state: SessionInstanceState): Promise<number> {
    // Count via the live blobs (not SCARD) so a stale id left behind by a crashed writer doesn't
    // inflate the cap-enforcement count; findByOrgTemplateState prunes dangling/stale index members
    // as a side effect.
    return (await this.findByOrgTemplateState(orgId, templateId, state)).length
  }

  async findByState(state: SessionInstanceState): Promise<SessionInstance[]> {
    const indexKey = SessionInstanceStore.STATE_IDX(state)
    const ids = await this.redis.smembers(indexKey)
    return this.readIndex(indexKey, ids, (inst) => inst.state === state)
  }

  async delete(id: string): Promise<void> {
    const prev = await this.findById(id)
    const pipe = this.redis.pipeline()
    pipe.del(SessionInstanceStore.KEY(id))
    if (prev) {
      pipe.srem(SessionInstanceStore.STATE_IDX(prev.state), id)
      pipe.srem(SessionInstanceStore.ORG_TPL_STATE_IDX(prev.organizationId, prev.templateId, prev.state), id)
    }
    await pipe.exec()
  }

  // -- internals ------------------------------------------------------------

  private async write(inst: SessionInstance, prev: SessionInstance | null): Promise<void> {
    const pipe = this.redis.pipeline()
    // Drop stale index membership when the state (or, defensively, org/template) changed.
    if (prev) {
      const stateChanged = prev.state !== inst.state
      const fleetChanged = prev.organizationId !== inst.organizationId || prev.templateId !== inst.templateId
      if (stateChanged || fleetChanged) {
        pipe.srem(SessionInstanceStore.STATE_IDX(prev.state), inst.id)
        pipe.srem(SessionInstanceStore.ORG_TPL_STATE_IDX(prev.organizationId, prev.templateId, prev.state), inst.id)
      }
    }
    pipe.set(SessionInstanceStore.KEY(inst.id), this.serialize(inst))
    pipe.sadd(SessionInstanceStore.STATE_IDX(inst.state), inst.id)
    pipe.sadd(SessionInstanceStore.ORG_TPL_STATE_IDX(inst.organizationId, inst.templateId, inst.state), inst.id)
    await pipe.exec()
  }

  /**
   * Resolve the ids of one index set to live instances, treating the blob as the
   * source of truth. Two kinds of bad entries are dropped from the result and
   * self-healed out of `indexKey`:
   *  - dangling: the blob is gone (crashed writer / external delete);
   *  - stale: the blob exists but no longer matches this index (a non-atomic
   *    concurrent save() can briefly leave an id in more than one state set).
   */
  private async readIndex(
    indexKey: string,
    ids: string[],
    matches: (inst: SessionInstance) => boolean,
  ): Promise<SessionInstance[]> {
    const out: SessionInstance[] = []
    const dangling: string[] = []
    const stale: string[] = []
    // Chunk MGET (MGET_CHUNK) so a large index set can't issue one unbounded
    // command / response — the same safeguard findByIds uses.
    for (let i = 0; i < ids.length; i += MGET_CHUNK) {
      const chunk = ids.slice(i, i + MGET_CHUNK)
      const raws = await this.redis.mget(chunk.map((id) => SessionInstanceStore.KEY(id)))
      raws.forEach((raw, j) => {
        if (!raw) {
          dangling.push(chunk[j])
          return
        }
        const inst = this.deserialize(raw)
        if (matches(inst)) out.push(inst)
        else stale.push(chunk[j])
      })
    }
    if (dangling.length > 0 || stale.length > 0) {
      this.pruneIndex(indexKey, dangling, stale).catch((err) =>
        this.logger.debug(`prune index ${indexKey} failed: ${err.message}`),
      )
    }
    return out
  }

  /**
   * Drop bad members discovered by readIndex. Stale members (blob exists but
   * doesn't match this index) are removed only from `indexKey`. Dangling members
   * (blob gone) are removed from `indexKey` AND swept out of every global
   * STATE_IDX, since an orphan can linger in multiple state sets and those keys
   * are enumerable (unlike the per-(org,template,state) sets, which is why the
   * dangling member must also be pruned from the specific `indexKey` here).
   */
  private async pruneIndex(indexKey: string, dangling: string[], stale: string[]): Promise<void> {
    const pipe = this.redis.pipeline()
    for (const id of stale) {
      pipe.srem(indexKey, id)
    }
    for (const id of dangling) {
      pipe.srem(indexKey, id)
      for (const state of Object.values(SessionInstanceState)) {
        pipe.srem(SessionInstanceStore.STATE_IDX(state), id)
      }
    }
    await pipe.exec()
  }

  private serialize(inst: SessionInstance): string {
    return JSON.stringify({
      id: inst.id,
      organizationId: inst.organizationId,
      templateId: inst.templateId,
      snapshotId: inst.snapshotId,
      sandboxId: inst.sandboxId,
      state: inst.state,
      errorReason: inst.errorReason,
      role: inst.role,
      lastUsedAt: inst.lastUsedAt?.toISOString(),
      lastActiveAt: inst.lastActiveAt?.toISOString(),
      createdAt: inst.createdAt.toISOString(),
      updatedAt: inst.updatedAt.toISOString(),
    })
  }

  private deserialize(raw: string): SessionInstance {
    const o = JSON.parse(raw)
    const inst = new SessionInstance()
    inst.id = o.id
    inst.organizationId = o.organizationId
    inst.templateId = o.templateId
    inst.snapshotId = o.snapshotId
    inst.sandboxId = o.sandboxId ?? undefined
    inst.state = o.state
    inst.errorReason = o.errorReason ?? undefined
    inst.role = o.role ?? SessionInstanceRole.WARM
    inst.lastUsedAt = o.lastUsedAt ? new Date(o.lastUsedAt) : undefined
    inst.lastActiveAt = o.lastActiveAt ? new Date(o.lastActiveAt) : undefined
    inst.createdAt = new Date(o.createdAt)
    inst.updatedAt = new Date(o.updatedAt)
    return inst
  }
}
