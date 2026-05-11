/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SessionInstanceStore } from './session-instance-store.service'
import { SessionInstanceState } from '../enums/session-instance-state.enum'
import { SessionInstanceRole } from '../enums/session-instance-role.enum'
import { FakeRedis } from './test-utils/fake-redis'

const ORG = 'org-1'
const TPL = 'tpl-1'

function newStore(): { store: SessionInstanceStore; redis: FakeRedis } {
  const redis = new FakeRedis()
  return { store: new SessionInstanceStore(redis as any), redis }
}

describe('SessionInstanceStore', () => {
  it('creates an instance discoverable by id, state, and org/template/state', async () => {
    const { store } = newStore()
    const inst = await store.create({ organizationId: ORG, templateId: TPL, role: SessionInstanceRole.WARM })

    expect(inst.id).toBeDefined()
    expect(inst.state).toBe(SessionInstanceState.PROVISIONING)
    expect(await store.findById(inst.id)).toMatchObject({ id: inst.id, organizationId: ORG })
    expect((await store.findByState(SessionInstanceState.PROVISIONING)).map((i) => i.id)).toEqual([inst.id])
    expect((await store.findByOrgTemplateState(ORG, TPL, SessionInstanceState.PROVISIONING)).map((i) => i.id)).toEqual([
      inst.id,
    ])
    expect(await store.countByState(ORG, TPL, SessionInstanceState.PROVISIONING)).toBe(1)
  })

  it('re-indexes state membership on a state transition', async () => {
    const { store } = newStore()
    const inst = await store.create({ organizationId: ORG, templateId: TPL })

    inst.state = SessionInstanceState.READY
    await store.save(inst)

    // Gone from the PROVISIONING indexes, present in the READY ones.
    expect(await store.findByState(SessionInstanceState.PROVISIONING)).toHaveLength(0)
    expect(await store.countByState(ORG, TPL, SessionInstanceState.PROVISIONING)).toBe(0)
    expect((await store.findByState(SessionInstanceState.READY)).map((i) => i.id)).toEqual([inst.id])
    expect(await store.countByState(ORG, TPL, SessionInstanceState.READY)).toBe(1)
  })

  it('update() patches fields without changing index membership', async () => {
    const { store } = newStore()
    const inst = await store.create({ organizationId: ORG, templateId: TPL, state: SessionInstanceState.READY })

    const when = new Date()
    await store.update(inst.id, { lastActiveAt: when })

    const reloaded = await store.findById(inst.id)
    expect(reloaded?.lastActiveAt?.getTime()).toBe(when.getTime())
    expect(await store.countByState(ORG, TPL, SessionInstanceState.READY)).toBe(1)
  })

  it('delete() removes the blob and all index entries', async () => {
    const { store } = newStore()
    const inst = await store.create({ organizationId: ORG, templateId: TPL, state: SessionInstanceState.READY })

    await store.delete(inst.id)

    expect(await store.findById(inst.id)).toBeNull()
    expect(await store.findByState(SessionInstanceState.READY)).toHaveLength(0)
    expect(await store.countByState(ORG, TPL, SessionInstanceState.READY)).toBe(0)
  })

  it('prunes a dangling index id whose blob has disappeared', async () => {
    const { store, redis } = newStore()
    // Simulate a crashed writer that left an id in the state index but no blob.
    await redis.sadd('session:inst:state:ready', 'ghost')

    expect(await store.findByState(SessionInstanceState.READY)).toHaveLength(0)
    // The dangling member was pruned as a side effect.
    expect(await redis.scard('session:inst:state:ready')).toBe(0)
  })

  it('prunes a dangling id from the per-org-template state index too, not just the global one', async () => {
    const { store, redis } = newStore()
    const orgTplKey = `session:inst:org:${ORG}:tpl:${TPL}:state:ready`
    await redis.sadd(orgTplKey, 'ghost')

    expect(await store.findByOrgTemplateState(ORG, TPL, SessionInstanceState.READY)).toHaveLength(0)
    expect(await redis.scard(orgTplKey)).toBe(0)
  })

  it('treats the blob as authoritative: a stale state-set entry is filtered out and self-healed', async () => {
    const { store, redis } = newStore()
    const inst = await store.create({ organizationId: ORG, templateId: TPL, state: SessionInstanceState.READY })
    // Simulate the residue of a non-atomic concurrent save(): the id is left in a
    // second (PROVISIONING) state set while the blob already says READY.
    const staleState = `session:inst:state:provisioning`
    const staleOrgTpl = `session:inst:org:${ORG}:tpl:${TPL}:state:provisioning`
    await redis.sadd(staleState, inst.id)
    await redis.sadd(staleOrgTpl, inst.id)

    // The stale PROVISIONING views must not surface the READY instance...
    expect(await store.findByState(SessionInstanceState.PROVISIONING)).toHaveLength(0)
    expect(await store.findByOrgTemplateState(ORG, TPL, SessionInstanceState.PROVISIONING)).toHaveLength(0)
    expect(await store.countByState(ORG, TPL, SessionInstanceState.PROVISIONING)).toBe(0)
    // ...and the stale membership is removed.
    expect(await redis.scard(staleState)).toBe(0)
    expect(await redis.scard(staleOrgTpl)).toBe(0)
    // The authoritative READY views are unaffected.
    expect((await store.findByState(SessionInstanceState.READY)).map((i) => i.id)).toEqual([inst.id])
    expect(await store.countByState(ORG, TPL, SessionInstanceState.READY)).toBe(1)
  })

  it('findByIds batch-loads instances and omits ids whose blob is missing', async () => {
    const { store } = newStore()
    const a = await store.create({ organizationId: ORG, templateId: TPL })
    const b = await store.create({ organizationId: ORG, templateId: TPL })

    const map = await store.findByIds([a.id, b.id, 'missing'])
    expect(map.size).toBe(2)
    expect(map.get(a.id)?.id).toBe(a.id)
    expect(map.get(b.id)?.organizationId).toBe(ORG)
    expect(map.get('missing')).toBeUndefined()

    // Empty input yields an empty map.
    expect((await store.findByIds([])).size).toBe(0)
  })

  it('findByIds chunks MGET across a large id list without dropping any', async () => {
    const { store } = newStore()
    const ids: string[] = []
    // 300 > MGET_CHUNK (256), so this spans multiple MGET batches.
    for (let i = 0; i < 300; i++) {
      const inst = await store.create({ organizationId: ORG, templateId: TPL })
      ids.push(inst.id)
    }

    const map = await store.findByIds(ids)
    expect(map.size).toBe(300)
    expect(map.get(ids[0])?.id).toBe(ids[0])
    expect(map.get(ids[255])?.id).toBe(ids[255]) // last of the first chunk
    expect(map.get(ids[256])?.id).toBe(ids[256]) // first of the second chunk
    expect(map.get(ids[299])?.id).toBe(ids[299])
  })

  it('findByState chunks its MGET across a large state index without dropping any', async () => {
    const { store } = newStore()
    // 300 > MGET_CHUNK (256), so readIndex spans multiple MGET batches.
    for (let i = 0; i < 300; i++) {
      await store.create({ organizationId: ORG, templateId: TPL })
    }
    expect(await store.findByState(SessionInstanceState.PROVISIONING)).toHaveLength(300)
  })
})
