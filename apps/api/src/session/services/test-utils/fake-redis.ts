/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { runRedisLua } from './redis-lua'

/**
 * Minimal in-memory Redis used by the session Redis-store unit tests. Implements only the
 * string / set / sorted-set / pipeline commands exercised by SessionInstanceStore,
 * SessionRepository and SessionGcService. Not a general-purpose fake.
 */
export class FakeRedis {
  strings = new Map<string, string>()
  sets = new Map<string, Set<string>>()
  zsets = new Map<string, Map<string, number>>()

  async get(key: string): Promise<string | null> {
    return this.strings.get(key) ?? null
  }

  async set(key: string, value: string): Promise<'OK'> {
    this.strings.set(key, value)
    return 'OK'
  }

  async del(...keys: string[]): Promise<number> {
    let n = 0
    for (const k of keys) {
      if (this.strings.delete(k) || this.sets.delete(k) || this.zsets.delete(k)) n++
    }
    return n
  }

  async mget(...args: Array<string | string[]>): Promise<(string | null)[]> {
    const keys = Array.isArray(args[0]) ? (args[0] as string[]) : (args as string[])
    return keys.map((k) => this.strings.get(k) ?? null)
  }

  async smembers(key: string): Promise<string[]> {
    return [...(this.sets.get(key) ?? [])]
  }

  async sadd(key: string, ...members: string[]): Promise<number> {
    const s = this.sets.get(key) ?? new Set<string>()
    let added = 0
    for (const m of members) {
      if (!s.has(m)) {
        s.add(m)
        added++
      }
    }
    this.sets.set(key, s)
    return added
  }

  async srem(key: string, ...members: string[]): Promise<number> {
    const s = this.sets.get(key)
    if (!s) return 0
    let n = 0
    for (const m of members) if (s.delete(m)) n++
    return n
  }

  async scard(key: string): Promise<number> {
    return this.sets.get(key)?.size ?? 0
  }

  async zadd(key: string, score: number | string, member: string): Promise<number> {
    const z = this.zsets.get(key) ?? new Map<string, number>()
    const had = z.has(member)
    z.set(member, Number(score))
    this.zsets.set(key, z)
    return had ? 0 : 1
  }

  async zrem(key: string, ...members: string[]): Promise<number> {
    const z = this.zsets.get(key)
    if (!z) return 0
    let n = 0
    for (const m of members) if (z.delete(m)) n++
    return n
  }

  async zrangebyscore(
    key: string,
    min: string | number,
    max: string | number,
    ...opts: (string | number)[]
  ): Promise<string[]> {
    const z = this.zsets.get(key)
    if (!z) return []
    const lo = min === '-inf' ? -Infinity : Number(min)
    const hi = max === '+inf' ? Infinity : Number(max)
    let arr = [...z.entries()]
      .filter(([, s]) => s >= lo && s <= hi)
      .sort((a, b) => a[1] - b[1])
      .map(([m]) => m)
    const li = opts.findIndex((o) => String(o).toUpperCase() === 'LIMIT')
    if (li >= 0) {
      const offset = Number(opts[li + 1])
      const count = Number(opts[li + 2])
      arr = arr.slice(offset, offset + count)
    }
    return arr
  }

  async zrevrange(key: string, start: number, stop: number): Promise<string[]> {
    const z = this.zsets.get(key)
    if (!z) return []
    const arr = [...z.entries()].sort((a, b) => b[1] - a[1]).map(([m]) => m)
    const end = stop === -1 ? arr.length : stop + 1
    return arr.slice(start, end)
  }

  pipeline(): FakePipeline {
    return new FakePipeline(this)
  }

  /**
   * Execute a real Redis Lua script through the fengari harness, routing
   * redis.call to this fake's own data (see luaCall). The session CAS scripts
   * (sweep / touch) therefore run verbatim in unit tests — there is no JS
   * reimplementation of their logic to drift from the production Lua.
   */
  async eval(script: string, numKeys: number, ...args: (string | number)[]): Promise<number | string | null> {
    const all = args.map(String)
    return runRedisLua(script, all.slice(0, numKeys), all.slice(numKeys), (cmd, a) => this.luaCall(cmd, a))
  }

  // redis.call binding for runRedisLua: the handful of commands the session CAS
  // scripts use, run synchronously against this fake's maps so the scripts and
  // the rest of the fake observe one consistent dataset. A missing key/score
  // returns `false`, matching Redis's Lua nil-reply convention.
  private luaCall(cmd: string, args: string[]): string | number | boolean | null {
    switch (cmd) {
      case 'GET': {
        const v = this.strings.get(args[0])
        return v === undefined ? false : v
      }
      case 'SET':
        this.strings.set(args[0], args[1])
        return 'OK'
      case 'ZSCORE': {
        const z = this.zsets.get(args[0])
        return z && z.has(args[1]) ? String(z.get(args[1])) : false
      }
      case 'ZADD': {
        let z = this.zsets.get(args[0])
        if (!z) {
          z = new Map<string, number>()
          this.zsets.set(args[0], z)
        }
        z.set(args[2], Number(args[1]))
        return 1
      }
      case 'ZREM': {
        const z = this.zsets.get(args[0])
        if (z) z.delete(args[1])
        return 1
      }
      default:
        throw new Error(`FakeRedis luaCall: unhandled command ${cmd}`)
    }
  }
}

/** Records chained commands and replays them against the backing FakeRedis on exec(). */
export class FakePipeline {
  private ops: Array<() => Promise<unknown>> = []

  constructor(private readonly redis: FakeRedis) {}

  set(...a: [string, string]): this {
    this.ops.push(() => this.redis.set(...a))
    return this
  }
  del(...a: string[]): this {
    this.ops.push(() => this.redis.del(...a))
    return this
  }
  sadd(...a: [string, ...string[]]): this {
    this.ops.push(() => this.redis.sadd(...a))
    return this
  }
  srem(...a: [string, ...string[]]): this {
    this.ops.push(() => this.redis.srem(...a))
    return this
  }
  zadd(...a: [string, number | string, string]): this {
    this.ops.push(() => this.redis.zadd(...a))
    return this
  }
  zrem(...a: [string, ...string[]]): this {
    this.ops.push(() => this.redis.zrem(...a))
    return this
  }
  eval(script: string, numKeys: number, ...args: (string | number)[]): this {
    this.ops.push(() => this.redis.eval(script, numKeys, ...args))
    return this
  }

  async exec(): Promise<Array<[Error | null, unknown]>> {
    const res: Array<[Error | null, unknown]> = []
    for (const op of this.ops) res.push([null, await op()])
    return res
  }
}
