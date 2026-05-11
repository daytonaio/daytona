// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0
//
// Daytona session-daemon — long-lived Node host that multiplexes many
// isolated-vm V8 sessions over stdin/stdout JSON-line protocol.
//
// Wire protocol:
//   in  (one line per cmd):  {"op":"create"|"exec"|"interrupt"|"delete"|"list-packages",
//                              "sessionId":"...","reply":"...?","code":"...?",
//                              "envs":{...?},"memoryLimitMb":N?,"execTimeoutMs":N?}
//   out (one line per chunk): {"sessionId":"...","type":"stdout"|"stderr"|"error"|
//                              "display"|"control","text":"..."?,
//                              "name":"..."?,"value":"..."?,"traceback":"..."?,
//                              "formats":[...]?,"data":{<mime>:...}?,
//                              "reply":"..."?,"packages":[...]?}

'use strict'

const path = require('node:path')
const readline = require('node:readline')
const fs = require('node:fs')
const fsp = require('node:fs/promises')

// Max bytes for any single text payload written into an output frame. A single
// frame is one stdout/stderr/display chunk on the wire; capping it here (the
// emit side) means no single line/frame can blow past the reader's bounded line
// size on the Go side, regardless of how much the user code prints in one shot.
const MAX_CHUNK_BYTES = 1 << 20 // 1 MiB

// Max bytes for a whole serialized output line (one JSON frame + newline). Kept
// safely below the Go reader's maxWorkerLineBytes (8 MiB) cap so even a frame
// with many text/data fields — each individually within MAX_CHUNK_BYTES — can
// never sum past the reader limit and trip its oversized-line recovery (which on
// the shared host drains-and-skips without notifying the affected exec, so a
// hung waiter is only avoided by keeping this path unreachable). Kept in sync
// with MAX_LINE_BYTES in repl_worker.py.
const MAX_LINE_BYTES = 4 << 20 // 4 MiB

// truncateText caps a UTF-8 text payload at MAX_CHUNK_BYTES, appending a clear
// marker reporting how many bytes were dropped. Returns the input untouched when
// it is already within budget (the common case).
function truncateText(s) {
  if (typeof s !== 'string') return s
  // Fast path: assume <= MAX_CHUNK_BYTES chars are <= MAX_CHUNK_BYTES bytes is
  // false for multibyte, so measure exactly when the char length is in range.
  const byteLen = Buffer.byteLength(s, 'utf8')
  if (byteLen <= MAX_CHUNK_BYTES) return s
  // Slice on a byte boundary, then back off to a valid UTF-8 boundary so we
  // never emit a split multibyte sequence.
  const buf = Buffer.from(s, 'utf8')
  let end = MAX_CHUNK_BYTES
  while (end > 0 && (buf[end] & 0xc0) === 0x80) end--
  const kept = buf.toString('utf8', 0, end)
  const omitted = byteLen - end
  return kept + `…[output truncated: ${omitted} bytes omitted]`
}

// shrinkFrame replaces a frame's text/data fields with a truncation marker.
// Last-resort fallback for emit() when an entire serialized frame would exceed
// MAX_LINE_BYTES (e.g. many fields each within MAX_CHUNK_BYTES but summing past
// the reader cap). Drops the bulky payloads so the line is small while keeping
// the frame's routing fields (sessionId/type/control text) intact. Mutates and
// returns the same object.
function shrinkFrame(chunk) {
  const marker = '…[output truncated: frame exceeded line limit]'
  if (typeof chunk.text === 'string') chunk.text = marker
  if (typeof chunk.value === 'string') chunk.value = marker
  if (typeof chunk.traceback === 'string') chunk.traceback = marker
  // `name` is user-controlled (e.g. an exception's constructor name), so a
  // pathological value must not survive into the shrunk frame.
  if (typeof chunk.name === 'string') chunk.name = marker
  if (chunk.data && typeof chunk.data === 'object') {
    for (const mime of Object.keys(chunk.data)) chunk.data[mime] = marker
  }
  return chunk
}

// stdoutMutex / emit (and writeChunkSync just below) are declared FIRST because
// the boot-error catch handlers below report through them before any other
// initialization runs. Reordering this triggers a temporal-dead-zone
// ReferenceError that masks the real boot error.
const stdoutMutex = (() => {
  let queue = Promise.resolve()
  return (fn) => {
    queue = queue.then(fn).catch(() => undefined)
    return queue
  }
})()

function emit(chunk) {
  // Cap oversized text payloads (stdout/stderr/error/display) before serializing
  // so no single frame can be unbounded — see MAX_CHUNK_BYTES. Truncation is
  // applied per text-bearing field rather than to the whole JSON line so the
  // structure stays intact and the omitted-bytes marker lands inside the field.
  if (chunk && typeof chunk === 'object') {
    if (typeof chunk.text === 'string') chunk.text = truncateText(chunk.text)
    if (typeof chunk.value === 'string') chunk.value = truncateText(chunk.value)
    if (typeof chunk.traceback === 'string') chunk.traceback = truncateText(chunk.traceback)
    if (chunk.data && typeof chunk.data === 'object') {
      for (const mime of Object.keys(chunk.data)) {
        if (typeof chunk.data[mime] === 'string') chunk.data[mime] = truncateText(chunk.data[mime])
      }
    }
  }
  // Single-writer through stdoutMutex so concurrent sessions don't interleave bytes.
  stdoutMutex(() => {
    try {
      let line = JSON.stringify(chunk)
      // Whole-line guard: per-field truncation above bounds each text/data field,
      // but a frame carrying many such fields could still serialize past the Go
      // reader's line cap. If so, drop the oversized text/data fields (replacing
      // them with a marker) so the emitted line is always under the reader limit
      // and the host's drain-and-skip oversized-line recovery stays unreachable.
      if (Buffer.byteLength(line, 'utf8') > MAX_LINE_BYTES) {
        line = JSON.stringify(shrinkFrame(chunk))
        // shrinkFrame only drops the known bulky fields (text/value/traceback/
        // name/data); a frame with other large fields (e.g. `packages` from
        // list-packages) could still exceed the cap. Re-check and, if still
        // oversized, fall back to a minimal control frame that preserves only
        // the routing fields so the emitted line is always under the reader cap.
        if (Buffer.byteLength(line, 'utf8') > MAX_LINE_BYTES) {
          line = JSON.stringify({
            sessionId: chunk.sessionId || '',
            type: chunk.type || 'control',
            text: '…[output truncated: frame exceeded line limit]',
          })
        }
      }
      process.stdout.write(line + '\n')
    } catch (e) {
      process.stderr.write('emit failed: ' + e.message + '\n')
    }
  })
}

// writeChunkSync writes a frame to stdout synchronously. The boot-error handlers
// below call process.exit(1) right after reporting, which terminates before
// emit()'s queued (stdoutMutex microtask) write to the async stdout pipe can
// flush — silently dropping the boot error. fs.writeSync delivers before exit.
function writeChunkSync(chunk) {
  try {
    fs.writeSync(1, JSON.stringify(chunk) + '\n')
  } catch (e) {
    try { process.stderr.write('writeChunkSync failed: ' + e.message + '\n') } catch (_e) { /* ignore */ }
  }
}

let ivm
try {
  // eslint-disable-next-line global-require
  ivm = require('isolated-vm')
} catch (err) {
  writeChunkSync({ type: 'error', name: 'HostBootError', value: 'isolated-vm not installed: ' + err.message })
  process.exit(1)
}

let esbuild
try {
  // eslint-disable-next-line global-require
  esbuild = require('esbuild-wasm')
} catch (err) {
  writeChunkSync({ type: 'error', name: 'HostBootError', value: 'esbuild-wasm not installed: ' + err.message })
  process.exit(1)
}

// just-bash backs the in-isolate bash() bridge. It is OPTIONAL: unlike
// isolated-vm/esbuild we do NOT exit on failure — TS sessions still run, and
// user code that calls bash() gets a clear "bash runtime unavailable" error.
// just-bash is "type":"module" but ships a CommonJS condition, so require()
// resolves it; OverlayFs is re-exported from the package root.
let Bash = null
let OverlayFs = null
try {
  // eslint-disable-next-line global-require
  const jb = require('just-bash')
  if (typeof jb.Bash === 'function' && typeof jb.OverlayFs === 'function') {
    Bash = jb.Bash
    OverlayFs = jb.OverlayFs
  }
} catch (_err) {
  // bash() bridge stays unavailable; logged lazily on first use.
}

const USER_NODE_MODULES_ROOT = process.env.SESSION_DAEMON_USER_NODE_MODULES_ROOT || '/workspace'

// Workspace root the in-isolate bash() bridge overlays (reads hit the real
// /workspace, writes stay private + in-memory per isolate). It is the same
// path the module resolver roots at.
const BASH_WORKSPACE_ROOT = process.env.SESSION_DAEMON_WORKSPACE_ROOT || USER_NODE_MODULES_ROOT

// Per-call wall-clock ceiling for bash() so a heavy/looping command can't stall
// the shared TS host event loop. Mirrors the bash host's own guard.
const BASH_CALL_TIMEOUT_MS = (() => {
  const n = parseInt(process.env.SESSION_DAEMON_BASH_EXEC_TIMEOUT_MS || '', 10)
  return Number.isFinite(n) && n > 0 ? n : 30000
})()

// just-bash execution-protection limits (runaway loops / deep recursion).
const BASH_EXECUTION_LIMITS = { maxCallDepth: 100, maxCommandCount: 100000, maxLoopIterations: 1000000 }

// One Map<sessionId, ContextRecord>.
const contexts = new Map()

// Bundle cache is per-host-process global, keyed on (packageName, version).
// Cache lifetime invariant — see plan §3 "module resolver + bundle cache".
const bundleCache = new Map()

class ContextRecord {
  constructor(opts) {
    this.id = opts.sessionId
    this.memoryLimitMb = opts.memoryLimitMb || 128
    this.isolate = new ivm.Isolate({ memoryLimit: this.memoryLimitMb })
    this.contextPromise = this.#bootContext()
    this.queue = Promise.resolve()
    this.disposed = false
    // Per-session compiled-Module cache. V8 modules cannot be shared across
    // sessions, so this cache lives on the ContextRecord (not in bundleCache,
    // which is the host-global source-string cache).
    this.modules = new Map()
    // Per-isolate just-bash instance backing the bash() bridge. Created lazily
    // on first bash() call and dropped on reset/dispose. Each isolate gets its
    // own OverlayFs so writes are private to that isolate and discarded on
    // teardown — same isolation posture as the standalone bash isolate.
    this.bash = null
  }

  // Lazily construct this isolate's bash shell. Throws if just-bash is
  // unavailable so the bridge surfaces a clear error to user code.
  #ensureBash() {
    if (!Bash || !OverlayFs) {
      throw makeError('BashUnavailableError', 'bash runtime unavailable')
    }
    if (!this.bash) {
      const overlay = new OverlayFs({ root: BASH_WORKSPACE_ROOT, mountPoint: BASH_WORKSPACE_ROOT })
      this.bash = new Bash({ fs: overlay, cwd: overlay.getMountPoint(), executionLimits: BASH_EXECUTION_LIMITS })
    }
    return this.bash
  }

  async #bootContext() {
    const ctx = await this.isolate.createContext()
    const jail = ctx.global
    await jail.set('global', jail.derefInto())

    // Console bridges → stdout/stderr chunks.
    await jail.set('_emitOut', new ivm.Reference((s) => emit({ sessionId: this.id, type: 'stdout', text: String(s) })))
    await jail.set('_emitErr', new ivm.Reference((s) => emit({ sessionId: this.id, type: 'stderr', text: String(s) })))

    // Fetch bridge (host-side fetch; session has no native networking).
    await jail.set('_fetch', new ivm.Reference(async (url, init) => {
      try {
        const initObj = init ? JSON.parse(init) : undefined
        const res = await fetch(url, initObj)
        const text = await res.text()
        return new ivm.ExternalCopy({
          ok: res.ok,
          status: res.status,
          statusText: res.statusText,
          url: res.url,
          headers: Object.fromEntries(res.headers.entries()),
          text,
        }).copyInto({ release: true })
      } catch (e) {
        return new ivm.ExternalCopy({ __error: true, message: e.message, name: e.name }).copyInto({ release: true })
      }
    }))

    // Bash bridge (host-side just-bash; session shells out to virtual bash
    // tooling). just-bash runs in the host JS engine, NOT inside the isolate
    // (isolated-vm has no Node), so we expose it via this Reference exactly like
    // the fetch bridge: only strings cross the boundary, and each isolate has
    // its own OverlayFs over /workspace (private, in-memory writes).
    await jail.set('_bash', new ivm.Reference(async (cmd, envStr) => {
      const controller = new AbortController()
      const timer = setTimeout(() => controller.abort(), BASH_CALL_TIMEOUT_MS)
      try {
        // Bound the env payload before parsing: JSON.parse is synchronous and a
        // pathologically large string would block the shared host event loop. The
        // env crosses from inside the memory-limited isolate, but a 1 MiB cap is
        // far above any legitimate use.
        if (typeof envStr === 'string' && Buffer.byteLength(envStr, 'utf8') > MAX_CHUNK_BYTES) {
          throw makeError('BashEnvTooLargeError', 'bash env payload exceeds size limit')
        }
        const env = envStr ? JSON.parse(envStr) : undefined
        const r = await this.#ensureBash().exec(String(cmd), { env, signal: controller.signal })
        // Cap stdout/stderr before copying into the isolate. just-bash output is
        // bounded only by execution limits (loop/command counts), so a noisy
        // command could otherwise pull an unbounded string into the shared daemon.
        // Same per-field MAX_CHUNK_BYTES cap the emit path applies.
        return new ivm.ExternalCopy({
          stdout: truncateText(r.stdout || ''),
          stderr: truncateText(r.stderr || ''),
          exitCode: typeof r.exitCode === 'number' ? r.exitCode : 0,
        }).copyInto({ release: true })
      } catch (e) {
        return new ivm.ExternalCopy({ __error: true, message: e.message, name: e.name }).copyInto({ release: true })
      } finally {
        clearTimeout(timer)
      }
    }))

    // Completion signaling for the async-IIFE wrapper used in #transformUserCode.
    // The wrapper calls __daytonaSignalOk on success or __daytonaSignalErr on
    // failure; per-exec we swap which Promise gets resolved/rejected by writing
    // the resolver pair into this._currentSignal. We need this side channel
    // because isolated-vm 5.x's module.evaluate({promise:true}) does NOT wait
    // for top-level await whose awaited promise crosses the host boundary
    // (notably the fetch shim) — see laverdet/isolated-vm#494. Without this
    // signal, the host would mark the exec "completed" before the user's TLA
    // actually ran.
    await jail.set('__daytonaSignalOk', new ivm.Reference(() => {
      const signal = this._currentSignal
      if (signal) signal.resolve()
    }))
    await jail.set('__daytonaSignalErr', new ivm.Reference((msg, name, stack) => {
      const signal = this._currentSignal
      if (!signal) return
      const e = new Error(typeof msg === 'string' ? msg : 'Error')
      if (typeof name === 'string' && name) e.name = name
      if (typeof stack === 'string' && stack) e.stack = stack
      signal.reject(e)
    }))

    // setTimeout/setInterval bridges (isolated-vm doesn't ship them). The
    // session calls these via applySync with `{ arguments: { reference: true } }`,
    // so `cb` arrives as an ivm.Reference we can invoke with applyIgnored, and
    // `ms` arrives as a Reference whose underlying number we copy out. Timers are
    // unref'd so a lingering user interval can't keep the host process (or, on
    // dispose, the isolate) alive past shutdown — clearInterval/clearTimeout and
    // ContextRecord.dispose still clear them explicitly.
    const timers = new Map()
    // Track timers on the instance so dispose() can clear any still-pending
    // user timers (notably intervals) instead of leaking them across a context
    // recycle (reset:true rebuilds the context but not the host process).
    this._timers = timers
    let nextTimerId = 1
    const derefMs = (ms) => {
      try { return Number(typeof ms === 'object' && ms ? ms.copySync() : ms) || 0 } catch { return 0 }
    }
    await jail.set('_setTimeout', new ivm.Reference((cb, ms) => {
      const id = nextTimerId++
      const handle = setTimeout(() => { timers.delete(id); try { cb.applyIgnored(undefined, []) } catch {} }, derefMs(ms))
      if (handle && typeof handle.unref === 'function') handle.unref()
      timers.set(id, handle)
      return id
    }))
    await jail.set('_setInterval', new ivm.Reference((cb, ms) => {
      const id = nextTimerId++
      const handle = setInterval(() => { try { cb.applyIgnored(undefined, []) } catch {} }, derefMs(ms))
      if (handle && typeof handle.unref === 'function') handle.unref()
      timers.set(id, handle)
      return id
    }))
    await jail.set('_clearTimer', new ivm.Reference((id) => {
      const handle = timers.get(id)
      if (!handle) return
      clearTimeout(handle)
      clearInterval(handle)
      timers.delete(id)
    }))

    // Install the user-facing globals (console, fetch, env, timers).
    const bootstrap = `
      const console = {
        log: (...args) => _emitOut.applySync(undefined, [args.map(_fmt).join(' ') + '\\n']),
        error: (...args) => _emitErr.applySync(undefined, [args.map(_fmt).join(' ') + '\\n']),
        warn: (...args) => _emitErr.applySync(undefined, [args.map(_fmt).join(' ') + '\\n']),
        info: (...args) => _emitOut.applySync(undefined, [args.map(_fmt).join(' ') + '\\n']),
        debug: (...args) => _emitOut.applySync(undefined, [args.map(_fmt).join(' ') + '\\n']),
      };
      function _fmt(v) {
        if (typeof v === 'string') return v;
        try { return JSON.stringify(v); } catch { return String(v); }
      }
      const setTimeout = (cb, ms) => _setTimeout.applySync(undefined, [cb, ms], { arguments: { reference: true } });
      const setInterval = (cb, ms) => _setInterval.applySync(undefined, [cb, ms], { arguments: { reference: true } });
      const clearTimeout = (id) => _clearTimer.applySync(undefined, [id]);
      const clearInterval = (id) => _clearTimer.applySync(undefined, [id]);
      // Minimal Headers shim — V8 sessions don't ship the Headers / Response /
      // Request Web APIs (those are Node-side only), so we provide a duck-typed
      // object that supports the .get / .has / .entries / .forEach methods user
      // code typically reaches for. Anything richer (e.g. setters) is
      // intentionally out of scope for this minimal shim.
      function _makeHeaders(raw) {
        const obj = raw || {};
        const lower = {};
        for (const k of Object.keys(obj)) lower[k.toLowerCase()] = obj[k];
        return {
          get(name) {
            const v = lower[String(name).toLowerCase()];
            return v == null ? null : v;
          },
          has(name) { return Object.prototype.hasOwnProperty.call(lower, String(name).toLowerCase()); },
          entries() { return Object.entries(obj); },
          keys() { return Object.keys(obj); },
          values() { return Object.values(obj); },
          forEach(cb, thisArg) {
            for (const k of Object.keys(obj)) cb.call(thisArg, obj[k], k, this);
          },
          [Symbol.iterator]() { return Object.entries(obj)[Symbol.iterator](); },
        };
      }
      const fetch = async (url, init) => {
        const initStr = init ? JSON.stringify(init) : undefined;
        const res = await _fetch.apply(undefined, [String(url), initStr], { result: { promise: true, copy: true } });
        if (res && res.__error) {
          throw new Error(res.message || 'fetch failed');
        }
        return {
          ok: res.ok, status: res.status, statusText: res.statusText, url: res.url,
          headers: _makeHeaders(res.headers || {}),
          text: async () => res.text,
          json: async () => JSON.parse(res.text),
          arrayBuffer: async () => new TextEncoder().encode(res.text).buffer,
        };
      };
      // bash(command[, env]) runs a command in this isolate's virtual just-bash
      // shell and resolves to { stdout, stderr, exitCode }. A non-zero exitCode
      // is NOT thrown — only a bridge/runtime failure rejects.
      const bash = async (command, env) => {
        const envStr = env ? JSON.stringify(env) : undefined;
        const res = await _bash.apply(undefined, [String(command), envStr], { result: { promise: true, copy: true } });
        if (res && res.__error) {
          throw new Error(res.message || 'bash failed');
        }
        return res;
      };
      globalThis.bash = bash;
      // env is exposed as a mutable plain global, populated per-call by the host
      // via ctx.global.set('env', ...). It cannot be declared with 'const' here
      // — that creates a lexical binding the host cannot override at exec time,
      // leaving env permanently empty.
      globalThis.env = {};
    `
    // The user-facing setTimeout/setInterval call the host bridges above with
    // `{ arguments: { reference: true } }`, so the callback crosses the boundary
    // as an ivm.Reference the bridge invokes — no `ivm` exposure to user code is
    // needed and the callbacks actually run.
    const compiled = await this.isolate.compileScript(bootstrap)
    await compiled.run(ctx)

    return ctx
  }

  async exec(code, envs, reset = false) {
    if (this.disposed) throw new Error('context disposed')
    return this.queue = this.queue.then(async () => {
      let ctx
      try {
        if (reset) {
          // Recycle the inner V8 context + per-session module cache so the
          // user code runs against pristine globals. Reusing the outer Session
          // (~3-5 ms) is what makes transient-context one-shot calls cheap
          // compared to a full ContextRecord rebuild. See ExecuteRequest.Reset.
          //
          // Clear the OLD context's timers FIRST: #bootContext reassigns
          // this._timers to a fresh Map, so without clearing here the previously
          // tracked intervals/timeouts keep firing against the discarded context
          // and leak host-side handles. See dispose() for the same teardown.
          this.#clearAllTimers()
          this.modules = new Map()
          // Drop the bash shell so the recycled context starts with a pristine
          // overlay (prior writes are discarded), matching the fresh-globals
          // semantics of a reset.
          this.bash = null
          this.contextPromise = this.#bootContext()
        }
        ctx = await this.contextPromise
      } catch (e) {
        emit({ sessionId: this.id, type: 'error', name: 'ContextBootError', value: e.message })
        emit({ sessionId: this.id, type: 'control', text: 'completed' })
        return
      }
      try {
        // Inject the per-call env values into the global `env` object.
        if (envs && typeof envs === 'object') {
          await ctx.global.set('env', new ivm.ExternalCopy(envs).copyInto({ release: true }))
        }
        // Reset the host-readable last-expression slot so a previous exec's
        // value can't leak into this one if the rewrite doesn't fire.
        await ctx.global.set(LAST_EXPR_GLOBAL, new ivm.ExternalCopy(undefined).copyInto({ release: true }))

        // Set up a host-side promise the IIFE wrapper signals into. We need
        // this because evaluate({promise:true}) doesn't reliably wait for
        // cross-boundary TLA (see #bootContext).
        const donePromise = new Promise((resolve, reject) => {
          this._currentSignal = { resolve, reject }
        })

        const userSource = await this.#transformUserCode(code)
        const userModule = await this.isolate.compileModule(userSource, { filename: 'user.mjs' })
        await userModule.instantiate(ctx, (spec, referrer) => this.#resolveModule(spec, referrer))
        // evaluate completes once the synchronous portion of the module body
        // finishes (the IIFE we emit returns immediately and runs in the
        // background). If the module body has a synchronous error — including
        // a TDZ error or a thrown error in import-side-effect code — evaluate
        // rejects and we surface it via the catch.
        await userModule.evaluate({ promise: true, timeout: 0 })
        // Wait for the async IIFE to call __daytonaSignalOk / __daytonaSignalErr.
        await donePromise

        // Last-expression value (rewritten by injectLastExprAsDefaultExport).
        const result = await ctx.global.get(LAST_EXPR_GLOBAL, { copy: true })
        if (result !== undefined) {
          this.#emitDisplayForValue(result)
        }
      } catch (e) {
        const err = unwrapError(e)
        emit({ sessionId: this.id, type: 'error', name: err.name, value: err.message, traceback: err.stack || '' })
      } finally {
        // Drop the resolver pair so any stray late call into the signal
        // bridges from a previous exec is a no-op (the bridges check this).
        this._currentSignal = null
        emit({ sessionId: this.id, type: 'control', text: 'completed' })
      }
    })
  }

  async #transformUserCode(code) {
    // Transpile TS/JS to ESM. We keep `import` syntax intact so V8 can link
    // it through isolated-vm's module loader (see ContextRecord.exec). The
    // older IIFE+banner/footer approach lowered `import` to `require()` and
    // forbade top-level await — both are fixed by emitting ESM here.
    const transformed = await esbuild.transform(code, {
      loader: 'ts',
      format: 'esm',
      target: 'es2022',
      supported: { 'top-level-await': true },
    })
    return injectLastExprAsDefaultExport(transformed.code)
  }

  async #ensureModule(spec) {
    // Resolve a bare specifier to a per-session compiled Module. The bundled
    // ESM source is cached host-process-globally in `bundleCache`; the
    // compiled Module must be per-session (V8 modules cannot be shared across
    // sessions) and is cached on `this.modules`.
    let m = this.modules.get(spec)
    if (m) return m
    let source = bundleCache.get(spec)
    if (source === undefined) {
      source = await this.#bundlePackage(spec)
      bundleCache.set(spec, source)
    }
    m = await this.isolate.compileModule(source, { filename: spec + '.mjs' })
    this.modules.set(spec, m)
    return m
  }

  // resolveCallback for Module.instantiate. ivm calls this for every transitive
  // dependency the linker encounters. Node builtins are surfaced as
  // NativeModuleError (the bundler kept them external precisely so we can
  // reject them here with a clean error name); relative specs are rejected
  // because contexts have no filesystem identity.
  async #resolveModule(specifier, _referrer) {
    if (NODE_BUILTINS.has(specifier)) {
      throw makeError('NativeModuleError',
        `Node builtin '${specifier}' is not available in TypeScript sessions. ` +
        `Use a pure-JS alternative or run via Python.`)
    }
    if (specifier.startsWith('.') || specifier.startsWith('/')) {
      throw makeError('ModuleNotFoundError', `Relative imports are not supported in sessions: ${specifier}`)
    }
    return this.#ensureModule(specifier)
  }

  async #bundlePackage(spec) {
    // Bundle a curated node_modules package into a single self-contained ESM
    // source string. The result is host-cacheable (per plan §3 "module resolver
    // + bundle cache") and consumed by session.compileModule for each session.
    //
    // mainFields/conditions prefer the package's ESM build when it ships dual
    // CJS+ESM (`"exports": { "import": "./esm/...", "require": "./cjs/..." }`).
    // Pure-CJS packages still bundle correctly — esbuild transpiles them to ESM
    // at bundle time. NODE_BUILTINS stay external so the resolveCallback can
    // surface them as NativeModuleError instead of bundling shims.
    const pkgRoot = path.join(USER_NODE_MODULES_ROOT, 'node_modules', spec)
    if (!fs.existsSync(pkgRoot)) {
      throw makeError('ModuleNotFoundError', `Cannot find package '${spec}' in ${USER_NODE_MODULES_ROOT}/node_modules`)
    }
    const pkgJsonPath = path.join(pkgRoot, 'package.json')
    if (fs.existsSync(pkgJsonPath)) {
      const pkgJson = JSON.parse(await fsp.readFile(pkgJsonPath, 'utf8'))
      if (await packageHasNativeBindings(pkgRoot, pkgJson)) {
        throw makeError('NativeModuleError',
          `Package '${spec}' requires native bindings, which are not supported in TypeScript sessions. ` +
          `Use a pure-JS alternative or run via Python.`)
      }
    }
    const result = await esbuild.build({
      entryPoints: [pkgRoot],
      bundle: true,
      write: false,
      format: 'esm',
      target: 'es2022',
      platform: 'neutral',
      mainFields: ['module', 'main'],
      conditions: ['import', 'default'],
      external: NODE_BUILTINS_LIST,
    })
    if (!result.outputFiles || result.outputFiles.length === 0) {
      throw makeError('BundleError', `esbuild produced no output for '${spec}'`)
    }
    return result.outputFiles[0].text
  }

  #emitDisplayForValue(value) {
    if (value === undefined || value === null) return

    const formats = []
    const data = {}
    if (typeof value === 'object' && typeof value.toHTML === 'function') {
      try {
        const html = String(value.toHTML())
        formats.push('text/html')
        data['text/html'] = html
      } catch (_e) {/* swallow */}
    }
    if (typeof value === 'object') {
      try {
        const json = JSON.stringify(value)
        formats.push('application/json')
        data['application/json'] = json
      } catch (_e) {/* circular structure → fall through */}
    }
    if (formats.length === 0) {
      formats.push('text/plain')
      data['text/plain'] = typeof value === 'string' ? value : String(value)
    }
    emit({ sessionId: this.id, type: 'display', formats, data })
  }

  // Clear every still-pending user timer (intervals especially) and empty the
  // tracking Map. Shared by dispose() and the reset path in exec() so a context
  // recycle doesn't leak timers that keep firing against a discarded context.
  #clearAllTimers() {
    if (!this._timers) return
    for (const handle of this._timers.values()) {
      clearTimeout(handle)
      clearInterval(handle)
    }
    this._timers.clear()
  }

  async dispose() {
    this.disposed = true
    // Clear any pending user timers so a leftover setInterval doesn't keep
    // firing (and leaking host-side handles) after the isolate is gone.
    this.#clearAllTimers()
    // The compiled Modules are owned by the session and are torn down with it;
    // we just drop our references. Explicit module.release() is unnecessary
    // because session.dispose() reclaims everything in one shot.
    this.modules.clear()
    // Drop the per-isolate bash shell + its overlay (in-memory writes discarded).
    this.bash = null
    try { this.isolate.dispose() } catch (_e) {/* already disposed */}
  }
}

function makeError(name, message) {
  const e = new Error(message)
  e.name = name
  return e
}

function unwrapError(e) {
  if (!e) return { name: 'UnknownError', message: 'unknown error', stack: '' }
  if (e.name && e.message) return e
  return { name: 'Error', message: String(e), stack: '' }
}

// Node builtin specifiers we keep `external` during package bundling so they
// surface to the resolveCallback (which rejects them as NativeModuleError)
// rather than getting silently shimmed in. Both the Set (O(1) membership) and
// the array (passed to esbuild's `external`) refer to the same source list.
const NODE_BUILTINS_LIST = [
  'fs', 'fs/promises', 'path', 'crypto', 'os', 'child_process',
  'http', 'https', 'net', 'tls', 'dns', 'stream', 'zlib', 'buffer',
  'url', 'util', 'events', 'querystring', 'assert', 'process',
  'node:fs', 'node:fs/promises', 'node:path', 'node:crypto', 'node:os',
  'node:child_process', 'node:http', 'node:https', 'node:net', 'node:tls',
  'node:dns', 'node:stream', 'node:zlib', 'node:buffer', 'node:url',
  'node:util', 'node:events', 'node:querystring', 'node:assert', 'node:process',
]
const NODE_BUILTINS = new Set(NODE_BUILTINS_LIST)

async function packageHasNativeBindings(pkgRoot, pkgJson) {
  if (pkgJson.binary || pkgJson.gypfile) return true
  const scripts = pkgJson.scripts || {}
  if (scripts.install && /node-gyp|prebuild|node-pre-gyp/.test(scripts.install)) return true
  const deps = Object.keys({ ...(pkgJson.dependencies || {}), ...(pkgJson.optionalDependencies || {}) })
  if (deps.some((d) => d === 'bindings' || d === 'node-gyp' || d === 'node-addon-api' || d === '@mapbox/node-pre-gyp')) {
    return true
  }
  // Walk the tree (shallow) for .node binaries.
  try {
    const entries = await fsp.readdir(pkgRoot, { withFileTypes: true })
    for (const entry of entries) {
      if (entry.isFile() && entry.name.endsWith('.node')) return true
      if (entry.isDirectory() && entry.name === 'build') {
        const sub = await fsp.readdir(path.join(pkgRoot, 'build'), { withFileTypes: true }).catch(() => [])
        if (sub.some((s) => s.name === 'Release' || s.name === 'Debug')) return true
      }
    }
  } catch (_e) {/* unreadable, assume safe */}
  return false
}

// Name of the globalThis slot the trailing-expression rewrite assigns to.
// We surface the value through globalThis (rather than a named export) so
// the host can read it via the already-live ctx.global Reference, which
// avoids a couple of isolated-vm rough edges around module-namespace lookups
// after TLA.
const LAST_EXPR_GLOBAL = '__daytonaLast'

function injectLastExprAsDefaultExport(esmCode) {
  // Two structural moves happen here:
  //
  //   1. Lift `import` / `export` declarations to module top — they MUST be
  //      at the module's top level by ESM grammar, so they cannot live inside
  //      the IIFE we emit below.
  //
  //   2. Wrap the rest of the user's code in an async IIFE that signals
  //      completion to the host via __daytonaSignalOk / __daytonaSignalErr.
  //      This is the workaround for laverdet/isolated-vm#494: in
  //      isolated-vm 5.x, module.evaluate({promise:true}) does NOT reliably
  //      wait for top-level await whose awaited promise crosses the host
  //      boundary (e.g. our fetch shim). Hoisting the user's awaits into an
  //      explicit async function makes the boundary obvious to ivm and the
  //      host-side signal fires only after the IIFE actually finishes.
  //
  // The trailing-expression rewrite is applied to the IIFE body so the value
  // ends up on globalThis BEFORE the success signal fires, which preserves
  // the Pyodide-style display UX from plan §4.
  const { top, body } = splitTopImportsExports(esmCode)
  const rewrittenBody = rewriteTrailingExpression(body)
  return top + '\n' + wrapBodyInAsyncIIFE(rewrittenBody) + '\n'
}

// Walks an ESM source and returns { top, body }, where `top` contains every
// leading import/export statement (including multi-line `import { … } from`)
// plus interleaved blank lines and line-comments, and `body` is everything
// from the first non-import/non-export statement onwards. The detector tracks
// brace depth across lines so multi-line `import { a, b } from "x"` doesn't
// get split midway.
function splitTopImportsExports(esmSource) {
  const lines = esmSource.split('\n')
  let i = 0
  let braceDepth = 0
  let inBlockComment = false
  while (i < lines.length) {
    const line = lines[i]
    if (inBlockComment) {
      if (line.indexOf('*/') !== -1) inBlockComment = false
      i++
      continue
    }
    if (braceDepth > 0) {
      braceDepth += (line.match(/\{/g) || []).length
      braceDepth -= (line.match(/\}/g) || []).length
      i++
      continue
    }
    const trimmed = line.trim()
    if (trimmed === '') { i++; continue }
    if (trimmed.startsWith('//')) { i++; continue }
    if (trimmed.startsWith('/*')) {
      if (trimmed.indexOf('*/') === -1) inBlockComment = true
      i++
      continue
    }
    if (/^(import|export)\b/.test(trimmed)) {
      const opens = (line.match(/\{/g) || []).length
      const closes = (line.match(/\}/g) || []).length
      if (opens > closes) braceDepth = opens - closes
      i++
      continue
    }
    break
  }
  return {
    top: lines.slice(0, i).join('\n'),
    body: lines.slice(i).join('\n'),
  }
}

function wrapBodyInAsyncIIFE(body) {
  // Note the leading semicolon: it defends against accidental ASI-fusion if
  // the preceding `top` happens to end with an expression that looks like a
  // call target (rare but cheap to prevent).
  return [
    ';(async () => {',
    '  try {',
    body,
    '  } catch (__daytona_err) {',
    '    const __m = __daytona_err && __daytona_err.message;',
    '    const __n = __daytona_err && __daytona_err.name;',
    '    const __s = __daytona_err && __daytona_err.stack;',
    '    __daytonaSignalErr.applyIgnored(undefined, [String(__m == null ? __daytona_err : __m), String(__n || "Error"), String(__s || "")]);',
    '    return;',
    '  }',
    '  __daytonaSignalOk.applyIgnored(undefined, []);',
    '})();',
  ].join('\n')
}

function rewriteTrailingExpression(jsCode) {
  // Find the final non-empty, non-comment line; if it looks like a bare
  // expression statement, rewrite it to `globalThis.<slot> = (expr);`. Lines
  // that start with declaration/control keywords are left untouched. AST-
  // correct rewrite is a v1.1 follow-up captured in plan §4.
  const assignTarget = 'globalThis[' + JSON.stringify(LAST_EXPR_GLOBAL) + ']'
  const lines = jsCode.split('\n')
  for (let i = lines.length - 1; i >= 0; i--) {
    const stripped = lines[i].trim()
    if (!stripped) continue
    if (stripped.startsWith('//') || stripped.startsWith('/*')) continue
    if (/^(let|const|var|function|class|import|export|return)\b/.test(stripped)) break
    if (/^if|^for|^while|^switch|^try|^throw/.test(stripped)) break
    if (stripped.endsWith(';') || stripped.endsWith('}')) {
      const expr = stripped.replace(/;$/, '')
      lines[i] = lines[i].replace(stripped, assignTarget + ' = (' + expr + ');')
    }
    break
  }
  return lines.join('\n')
}

async function listPackages() {
  const root = path.join(USER_NODE_MODULES_ROOT, 'node_modules')
  if (!fs.existsSync(root)) return []
  const out = []
  const entries = await fsp.readdir(root, { withFileTypes: true })
  for (const entry of entries) {
    if (!entry.isDirectory()) continue
    if (entry.name.startsWith('.')) continue
    if (entry.name.startsWith('@')) {
      const subEntries = await fsp.readdir(path.join(root, entry.name), { withFileTypes: true }).catch(() => [])
      for (const sub of subEntries) {
        if (!sub.isDirectory()) continue
        const fullName = entry.name + '/' + sub.name
        out.push(await packageInfo(root, fullName))
      }
    } else {
      out.push(await packageInfo(root, entry.name))
    }
  }
  return out.filter(Boolean)
}

async function packageInfo(root, name) {
  const pkgRoot = path.join(root, name)
  const pkgJsonPath = path.join(pkgRoot, 'package.json')
  if (!fs.existsSync(pkgJsonPath)) return null
  try {
    const pkgJson = JSON.parse(await fsp.readFile(pkgJsonPath, 'utf8'))
    return {
      name,
      version: pkgJson.version || '',
      hasNativeBindings: await packageHasNativeBindings(pkgRoot, pkgJson),
    }
  } catch (_e) {
    return { name, version: '', hasNativeBindings: false }
  }
}

// ---------------------- Stdin command loop ----------------------

const rl = readline.createInterface({ input: process.stdin })
rl.on('line', async (line) => {
  let cmd
  try { cmd = JSON.parse(line) } catch (e) {
    emit({ type: 'error', name: 'JSONParseError', value: e.message })
    return
  }
  try {
    switch (cmd.op) {
      case 'create': {
        if (contexts.has(cmd.sessionId)) {
          emit({ sessionId: cmd.sessionId, type: 'error', name: 'ContextExistsError', value: 'context already exists' })
          return
        }
        contexts.set(cmd.sessionId, new ContextRecord(cmd))
        emit({ sessionId: cmd.sessionId, type: 'control', text: 'created' })
        break
      }
      case 'exec': {
        const ctx = contexts.get(cmd.sessionId)
        if (!ctx) {
          emit({ sessionId: cmd.sessionId, type: 'error', name: 'ContextNotFoundError', value: 'context not found' })
          emit({ sessionId: cmd.sessionId, type: 'control', text: 'completed' })
          return
        }
        // Fire-and-forget; the contextRecord queues internally so concurrent
        // exec frames for one context process FIFO. `reset:true` is the
        // transient-context recycling flag — see ExecuteRequest in types.go.
        ctx.exec(cmd.code, cmd.envs, !!cmd.reset)
        break
      }
      case 'interrupt': {
        const ctx = contexts.get(cmd.sessionId)
        if (!ctx) return
        await ctx.dispose()
        contexts.set(cmd.sessionId, new ContextRecord({ sessionId: cmd.sessionId, memoryLimitMb: ctx.memoryLimitMb }))
        emit({ sessionId: cmd.sessionId, type: 'control', text: 'interrupted' })
        break
      }
      case 'delete': {
        const ctx = contexts.get(cmd.sessionId)
        if (!ctx) return
        await ctx.dispose()
        contexts.delete(cmd.sessionId)
        emit({ sessionId: cmd.sessionId, type: 'control', text: 'deleted' })
        break
      }
      case 'list-packages': {
        const pkgs = await listPackages()
        emit({ type: 'control', text: 'list-packages-result', reply: cmd.reply || '', packages: pkgs })
        break
      }
      default:
        emit({ sessionId: cmd.sessionId || '', type: 'error', name: 'UnknownOpError', value: 'unknown op: ' + cmd.op })
    }
  } catch (e) {
    emit({ sessionId: cmd.sessionId || '', type: 'error', name: 'HostError', value: e.message })
  }
})

rl.on('close', () => {
  for (const ctx of contexts.values()) {
    try { ctx.dispose() } catch {}
  }
  process.exit(0)
})

// Surface the host's lifecycle to the Go side.
emit({ type: 'control', text: 'host-ready' })
