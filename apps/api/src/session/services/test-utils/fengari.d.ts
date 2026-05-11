/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

// fengari (a pure-JS Lua VM, used only in tests to run the real Redis Lua scripts)
// ships no TypeScript types. Declaring the module as untyped lets the test-only
// harness in redis-lua.ts import its C-style API without per-call annotations.
declare module 'fengari'
