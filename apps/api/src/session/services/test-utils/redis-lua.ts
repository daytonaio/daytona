/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Runs a real Redis Lua script in unit tests via fengari (a pure-JS Lua VM). It
 * recreates the environment Redis gives a script — `redis.call` (routed to the
 * supplied callback), `cjson` (backed by JSON), and the `KEYS`/`ARGV` arrays — so
 * the ACTUAL script text executes. Crucially there is NO JS reimplementation of
 * the script's logic to drift from the real Lua; only Redis's surrounding
 * environment is emulated, and that emulation is generic (script-independent).
 * Used by FakeRedis.eval so the session CAS scripts are exercised verbatim.
 */
import { lua, lauxlib, lualib, to_luastring } from 'fengari'

/**
 * redis.call binding: receives the command name (upper-cased) and its string
 * args, returns the reply — a string for a bulk reply, a number for an integer
 * reply, or `false` for a nil/missing reply (matching Redis's Lua conventions,
 * where a missing key surfaces as `false`, not `nil`).
 */
export type RedisCall = (cmd: string, args: string[]) => string | number | boolean | null

/** Execute `script` with the given KEYS/ARGV, routing redis.call to `call`. */
export function runRedisLua(script: string, keys: string[], argv: string[], call: RedisCall): number | string | null {
  const L = lauxlib.luaL_newstate()
  try {
    lualib.luaL_openlibs(L)
    installRedis(L, call)
    installCjson(L)
    installArray(L, 'KEYS', keys)
    installArray(L, 'ARGV', argv)

    const status = lauxlib.luaL_dostring(L, to_luastring(script))
    if (status !== lua.LUA_OK) {
      throw new Error('runRedisLua: ' + lua.lua_tojsstring(L, -1))
    }
    // Our scripts return a single integer (0/1); tolerate a string reply too.
    if (lua.lua_isnumber(L, -1)) return lua.lua_tonumber(L, -1)
    if (lua.lua_isstring(L, -1)) return lua.lua_tojsstring(L, -1)
    return null
  } finally {
    // luaL_newstate allocates a full Lua VM (stdlib, registry, string table);
    // free it so each eval() in tests doesn't leak an entire interpreter state.
    lua.lua_close(L)
  }
}

function installRedis(L: any, call: RedisCall): void {
  lua.lua_newtable(L)
  lua.lua_pushjsfunction(L, (Ls: any) => {
    const n = lua.lua_gettop(Ls)
    const args: string[] = []
    for (let i = 1; i <= n; i++) {
      args.push(lua.lua_type(Ls, i) === lua.LUA_TNUMBER ? String(lua.lua_tonumber(Ls, i)) : lua.lua_tojsstring(Ls, i))
    }
    pushReply(Ls, call(String(args[0]).toUpperCase(), args.slice(1)))
    return 1
  })
  lua.lua_setfield(L, -2, to_luastring('call'))
  lua.lua_setglobal(L, to_luastring('redis'))
}

function installCjson(L: any): void {
  lua.lua_newtable(L)
  lua.lua_pushjsfunction(L, (Ls: any) => {
    let obj: Record<string, unknown>
    try {
      obj = JSON.parse(lua.lua_tojsstring(Ls, 1))
    } catch (e) {
      // Mirror real cjson.decode: raise a Lua error (catchable by pcall) on bad JSON.
      return lauxlib.luaL_error(Ls, to_luastring('cjson.decode: ' + String(e)))
    }
    pushObject(Ls, obj)
    return 1
  })
  lua.lua_setfield(L, -2, to_luastring('decode'))
  lua.lua_pushjsfunction(L, (Ls: any) => {
    lua.lua_pushstring(Ls, to_luastring(JSON.stringify(readObject(Ls, 1))))
    return 1
  })
  lua.lua_setfield(L, -2, to_luastring('encode'))
  lua.lua_setglobal(L, to_luastring('cjson'))
}

function pushReply(L: any, v: string | number | boolean | null): void {
  if (v === false || v === null || v === undefined) lua.lua_pushboolean(L, false)
  else if (typeof v === 'number') lua.lua_pushnumber(L, v)
  else lua.lua_pushstring(L, to_luastring(String(v)))
}

// Push a flat JS object as a Lua table (the shapes our cjson.decode handles —
// the session context/instance blobs are flat string/number maps).
function pushObject(L: any, obj: Record<string, unknown>): void {
  lua.lua_newtable(L)
  for (const k of Object.keys(obj)) {
    const val = obj[k]
    if (val === null || val === undefined) continue
    lua.lua_pushstring(L, to_luastring(k))
    if (typeof val === 'number') lua.lua_pushnumber(L, val)
    else if (typeof val === 'boolean') lua.lua_pushboolean(L, val)
    else lua.lua_pushstring(L, to_luastring(String(val)))
    lua.lua_settable(L, -3)
  }
}

// Read a Lua table at `idx` back into a flat JS object (for cjson.encode).
function readObject(L: any, idx: number): Record<string, unknown> {
  const t = lua.lua_absindex(L, idx)
  const obj: Record<string, unknown> = {}
  lua.lua_pushnil(L)
  while (lua.lua_next(L, t) !== 0) {
    const key = lua.lua_tojsstring(L, -2)
    const ty = lua.lua_type(L, -1)
    if (ty === lua.LUA_TNUMBER) obj[key] = lua.lua_tonumber(L, -1)
    else if (ty === lua.LUA_TBOOLEAN) obj[key] = lua.lua_toboolean(L, -1)
    else obj[key] = lua.lua_tojsstring(L, -1)
    lua.lua_pop(L, 1)
  }
  return obj
}

function installArray(L: any, name: string, arr: string[]): void {
  lua.lua_newtable(L)
  arr.forEach((v, i) => {
    lua.lua_pushstring(L, to_luastring(String(v)))
    lua.lua_rawseti(L, -2, i + 1)
  })
  lua.lua_setglobal(L, to_luastring(name))
}
