/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SessionPoolService } from './session-pool.service'
import { SessionInstance } from '../entities/session-instance.entity'
import { SessionInstanceState } from '../enums/session-instance-state.enum'
import { SessionInstanceRole } from '../enums/session-instance-role.enum'
import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { TypedConfigService } from '../../config/typed-config.service'

const ORG = 'org-1'
const TPL = 'tpl-1'

function instance(opts: Partial<SessionInstance> & { id: string }): SessionInstance {
  const e = new SessionInstance()
  e.organizationId = ORG
  e.templateId = TPL
  e.snapshotId = 'snap-1'
  e.state = SessionInstanceState.READY
  e.role = SessionInstanceRole.OVERFLOW
  e.sandboxId = `sbx-${opts.id}`
  Object.assign(e, opts)
  return e
}

function makeConfig(overrides: Record<string, number> = {}): TypedConfigService {
  const map: Record<string, number> = {
    'session.scale.minWarm': 1,
    'session.scale.scaleInIdleSeconds': 600,
    ...overrides,
  }
  return { get: (key: string) => map[key] } as unknown as TypedConfigService
}

/**
 * Build a SessionPoolService with just the dependencies scale-in / prune touch mocked.
 * `seedInstances` is the full set the SessionInstanceStore holds; `findByState` filters it by
 * state so scale-in (READY) and prune (ERROR) each see the right slice. `loads` maps
 * instanceId -> effective load.
 */
function makePool(seedInstances: SessionInstance[], loads: Record<string, number> = {}) {
  const instances = {
    findByState: jest.fn(async (state: SessionInstanceState) => seedInstances.filter((i) => i.state === state)),
    findById: jest.fn(async (id: string) => seedInstances.find((i) => i.id === id) ?? null),
    save: jest.fn(async (x: SessionInstance) => x),
    update: jest.fn(async () => undefined),
    delete: jest.fn(async () => undefined),
    create: jest.fn(async (x: any) => x),
    findByOrgTemplateState: jest.fn(async () => []),
    countByState: jest.fn(async () => 0),
  }
  const sandboxService = { destroy: jest.fn(async () => undefined) }
  const sessions = { markInstanceSessionsInvalid: jest.fn(async () => undefined) }
  const load = { effectiveLoad: jest.fn(async (id: string) => loads[id] ?? 0) }
  // rollInstance now destroys the live sandbox; the repo answers with a STARTED row per instance.
  const sandboxRepo = {
    findOne: jest.fn(async ({ where: { id } }: any) => ({ id, state: SandboxState.STARTED })),
  }

  const pool = new SessionPoolService(
    instances as any,
    {} as any, // templateRepo
    sandboxRepo as any,
    sandboxService as any,
    sessions as any,
    {} as any, // lockProvider
    makeConfig() as any,
    load as any,
    {} as any, // scheduler
  )
  return { pool, instances, sandboxService, sessions, load, sandboxRepo }
}

function withConfig(pool: SessionPoolService, cfg: TypedConfigService): void {
  // Swap the private config so individual tests can tune minWarm / idle.
  ;(pool as any).config = cfg
}

const old = () => new Date(Date.now() - 700_000) // > 600s idle
const recent = () => new Date(Date.now() - 10_000) // < 600s

describe('SessionPoolService scale-in', () => {
  describe('scaleIn', () => {
    it('reaps only idle, zero-load overflow instances and leaves warm/loaded/active ones', async () => {
      const warm = instance({ id: 'w', role: SessionInstanceRole.WARM, lastActiveAt: old() })
      const idleFree = instance({ id: 'o1', lastActiveAt: old() }) // idle + load 0 -> reap
      const idleBusy = instance({ id: 'o2', lastActiveAt: old() }) // idle but has load -> keep
      const activeFree = instance({ id: 'o3', lastActiveAt: recent() }) // load 0 but not idle -> keep

      const { pool, sandboxService, sessions, instances } = makePool([warm, idleFree, idleBusy, activeFree], {
        o1: 0,
        o2: 3,
        o3: 0,
      })

      await (pool as any).scaleIn()

      expect(sandboxService.destroy).toHaveBeenCalledTimes(1)
      expect(sandboxService.destroy).toHaveBeenCalledWith('sbx-o1', ORG)
      expect(sessions.markInstanceSessionsInvalid).toHaveBeenCalledWith('o1')
      // The reaped instance row was marked ERROR before destroy.
      const saved = (instances.save as jest.Mock).mock.calls.map((c) => c[0] as SessionInstance)
      expect(saved).toHaveLength(1)
      expect(saved[0].id).toBe('o1')
      expect(saved[0].state).toBe(SessionInstanceState.ERROR)
    })

    it('never reaps below minWarm and removes the oldest-idle overflow first', async () => {
      // Two overflow instances, no warm, minWarm=1 -> exactly one reaped (the older).
      const older = instance({ id: 'older', lastActiveAt: new Date(Date.now() - 900_000) })
      const newer = instance({ id: 'newer', lastActiveAt: new Date(Date.now() - 700_000) })

      const { pool, sandboxService } = makePool([older, newer], { older: 0, newer: 0 })
      await (pool as any).scaleIn()

      expect(sandboxService.destroy).toHaveBeenCalledTimes(1)
      expect(sandboxService.destroy).toHaveBeenCalledWith('sbx-older', ORG)
    })

    it('does nothing when the fleet is already at or below minWarm', async () => {
      const warm = instance({ id: 'w', role: SessionInstanceRole.WARM, lastActiveAt: old() })
      const overflow = instance({ id: 'o1', lastActiveAt: old() })

      const { pool, sandboxService, sessions } = makePool([warm, overflow], { o1: 0 })
      withConfig(pool, makeConfig({ 'session.scale.minWarm': 2 }))

      await (pool as any).scaleIn()

      expect(sandboxService.destroy).not.toHaveBeenCalled()
      expect(sessions.markInstanceSessionsInvalid).not.toHaveBeenCalled()
    })

    it('does not reap an idle overflow that still reports load (e.g. an active SDK stream)', async () => {
      const a = instance({ id: 'o1', lastActiveAt: old() })
      const b = instance({ id: 'o2', lastActiveAt: old() })
      // Both idle by clock, but both currently serving -> nothing safe to reap.
      const { pool, sandboxService } = makePool([a, b], { o1: 1, o2: 2 })

      await (pool as any).scaleIn()
      expect(sandboxService.destroy).not.toHaveBeenCalled()
    })

    it('isolates fleets per (org, template)', async () => {
      const a1 = instance({ id: 'a1', lastActiveAt: old() })
      const a2 = instance({ id: 'a2', lastActiveAt: old() })
      // A different template with a single instance must never be touched.
      const b1 = instance({ id: 'b1', templateId: 'tpl-2', sandboxId: 'sbx-b1', lastActiveAt: old() })

      const { pool, sandboxService } = makePool([a1, a2, b1], { a1: 0, a2: 0, b1: 0 })
      await (pool as any).scaleIn()

      // tpl-1 had 2 -> reap 1; tpl-2 had 1 (== minWarm) -> untouched.
      expect(sandboxService.destroy).toHaveBeenCalledTimes(1)
      expect(sandboxService.destroy).not.toHaveBeenCalledWith('sbx-b1', ORG)
    })
  })

  describe('pruneErroredInstances', () => {
    it('deletes only ERROR rows older than the grace cutoff', async () => {
      // grace = scaleInIdleSeconds (600s). 'stale' aged out; 'fresh' is within grace.
      const stale = instance({
        id: 'stale',
        state: SessionInstanceState.ERROR,
        updatedAt: new Date(Date.now() - 700_000),
      })
      const fresh = instance({
        id: 'fresh',
        state: SessionInstanceState.ERROR,
        updatedAt: new Date(Date.now() - 10_000),
      })
      const { pool, instances } = makePool([stale, fresh])

      await (pool as any).pruneErroredInstances()

      expect(instances.delete).toHaveBeenCalledTimes(1)
      expect(instances.delete).toHaveBeenCalledWith('stale')
    })
  })
})
