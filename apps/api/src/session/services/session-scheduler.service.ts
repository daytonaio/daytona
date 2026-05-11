/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { SessionInstance } from '../entities/session-instance.entity'
import { SessionLoadService } from './session-load.service'
import { TypedConfigService } from '../../config/typed-config.service'

/**
 * SessionScheduler decides which warm sandbox a new (context-less) request should land on.
 *
 * It is a pure selector — it never provisions. The pool owns provisioning and calls
 * `claim`/`claimForce` to pick among existing READY instances. Selection is an *atomic claim*:
 * it `incr`s the instance's in-flight counter and keeps the increment only if the post-increment
 * load is within budget. This is what prevents a simultaneous burst from all reading "load 0" and
 * stampeding onto the same sandbox — each concurrent claimer sees a strictly higher counter.
 *
 * The caller (SessionService) MUST release the claimed slot with `SessionLoadService.decrInflight`
 * once its op (or initial setup, for persistent sessions) completes.
 */
@Injectable()
export class SessionScheduler {
  constructor(
    private readonly load: SessionLoadService,
    private readonly config: TypedConfigService,
  ) {}

  /**
   * Atomically claim a slot on the least-loaded instance that still has headroom (below the
   * per-sandbox concurrency target and not under resource pressure). Returns the claimed
   * instance, or null if every instance is saturated.
   */
  async claim(instances: SessionInstance[]): Promise<SessionInstance | null> {
    if (instances.length === 0) return null
    const target = this.config.get('session.scale.targetConcurrencyPerSandbox') ?? 4

    for (const inst of await this.byAscendingLoad(instances)) {
      const n = await this.load.incrInflight(inst.id)
      // A failed increment (Redis down) returns a negative sentinel — fail closed: don't treat it
      // as a free slot, and don't decrement (nothing was incremented). Try the next instance.
      if (n < 0) continue
      const snap = await this.load.getSnapshot(inst.id)
      const effective = Math.max(n, snap?.busyContexts ?? 0)
      if (effective <= target && !this.load.isResourceSaturated(snap)) {
        return inst
      }
      // Over budget / under pressure — release the optimistic claim and try the next instance.
      await this.load.decrInflight(inst.id)
    }
    return null
  }

  /**
   * Overload fallback: claim the least-loaded instance regardless of saturation. Used when the
   * fleet is at its instance cap and every sandbox is already full, so the request is better
   * served (queued behind a busy context) than rejected.
   */
  async claimForce(instances: SessionInstance[]): Promise<SessionInstance | null> {
    const ordered = await this.byAscendingLoad(instances)
    const chosen = ordered[0]
    if (!chosen) return null
    await this.load.incrInflight(chosen.id)
    return chosen
  }

  private async byAscendingLoad(instances: SessionInstance[]): Promise<SessionInstance[]> {
    const scored = await Promise.all(
      instances.map(async (inst) => ({ inst, load: await this.load.effectiveLoad(inst.id) })),
    )
    scored.sort((a, b) => a.load - b.load)
    return scored.map((s) => s.inst)
  }
}
