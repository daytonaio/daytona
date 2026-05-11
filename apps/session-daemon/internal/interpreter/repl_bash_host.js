// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0
//
// Daytona session-daemon — long-lived Node host that multiplexes many
// just-bash virtual shells over a stdin/stdout JSON-line protocol. This is the
// bash-isolate engine: each session gets its own `Bash` instance with an
// OverlayFs over the real /workspace (reads hit disk, writes stay in memory and
// are discarded on teardown). There are NO real subprocesses and NO real
// binaries — just-bash reimplements grep/sed/awk/jq/find/pipes/etc. in-process.
//
// Wire protocol:
//   in  (one line per cmd):  {"op":"create"|"exec"|"interrupt"|"delete"|"bash-call",
//                              "sessionId":"...","reply":"...?","code":"...?",
//                              "envs":{...?},"execTimeoutMs":N?,"cwd":"...?"}
//   out (one line per chunk): {"sessionId":"...","type":"stdout"|"stderr"|"error"|
//                              "control","text":"..."?,"name":"..."?,"value":"..."?}
//
// `exec` is the streaming path used by the standalone bash isolate: it emits
// stdout/stderr chunks then a terminal {type:"control",text:"completed"}.
// `bash-call` is the one-shot reply-correlated path used by the TS/Python
// bash() bridges: it returns a single {type:"control",text:"bash-call-result",
// reply,stdout,stderr,exitCode} (or text:"bash-call-error").

'use strict'

const readline = require('node:readline')
const fs = require('node:fs')

const stdoutMutex = (() => {
  let queue = Promise.resolve()
  return (fn) => {
    queue = queue.then(fn).catch(() => undefined)
    return queue
  }
})()

function emit(chunk) {
  stdoutMutex(() => {
    try {
      process.stdout.write(JSON.stringify(chunk) + '\n')
    } catch (e) {
      process.stderr.write('emit failed: ' + e.message + '\n')
    }
  })
}

// writeChunkSync writes one frame to stdout synchronously. emit() queues the
// write through stdoutMutex (a microtask) and process.stdout is an async pipe, so
// on a path that calls process.exit() immediately afterwards (the boot-error
// handler) the queued write never flushes and the frame is silently dropped.
// fs.writeSync guarantees the bytes reach fd 1 before the process exits.
function writeChunkSync(chunk) {
  try {
    fs.writeSync(1, JSON.stringify(chunk) + '\n')
  } catch (e) {
    try { process.stderr.write('writeChunkSync failed: ' + e.message + '\n') } catch (_e) { /* ignore */ }
  }
}

let Bash
let OverlayFs
try {
  // just-bash is "type":"module" but ships a CommonJS condition
  // (dist/bundle/index.cjs); require() resolves it here. OverlayFs is
  // re-exported from the package root, so the unexported "just-bash/fs/*"
  // subpaths are intentionally avoided.
  // eslint-disable-next-line global-require
  const jb = require('just-bash')
  Bash = jb.Bash
  OverlayFs = jb.OverlayFs
  if (typeof Bash !== 'function' || typeof OverlayFs !== 'function') {
    throw new Error('just-bash did not export Bash/OverlayFs')
  }
} catch (err) {
  // Synchronous write (not emit): process.exit(1) terminates before emit()'s
  // queued microtask write flushes, which would drop this boot error.
  writeChunkSync({ type: 'error', name: 'HostBootError', value: 'just-bash not installed: ' + err.message })
  process.exit(1)
}

const WORKSPACE_ROOT = process.env.SESSION_DAEMON_WORKSPACE_ROOT || '/workspace'

// Per-call wall-clock ceiling for the bash() bridges and a default for the
// streaming exec path when the caller didn't pass an execTimeoutMs.
const DEFAULT_EXEC_TIMEOUT_MS = intFromEnv('SESSION_DAEMON_BASH_EXEC_TIMEOUT_MS', 30000)

// just-bash execution-protection limits (guard against runaway loops/recursion
// that would otherwise spin the shared host). All have library defaults; we
// surface the common ones so they can be tuned per-deployment.
const EXECUTION_LIMITS = {
  maxCallDepth: intFromEnv('SESSION_DAEMON_BASH_MAX_CALL_DEPTH', 100),
  maxCommandCount: intFromEnv('SESSION_DAEMON_BASH_MAX_COMMAND_COUNT', 100000),
  maxLoopIterations: intFromEnv('SESSION_DAEMON_BASH_MAX_LOOP_ITERATIONS', 1000000),
}

function intFromEnv(name, dflt) {
  const raw = process.env[name]
  if (!raw) return dflt
  const n = parseInt(raw, 10)
  return Number.isFinite(n) && n > 0 ? n : dflt
}

// Distinct abort reason for a per-call timeout so run() can tell a timeout apart
// from an explicit interrupt via controller.signal.reason. AbortController.abort()
// is idempotent (the first call records the reason), which makes the
// classification race-free: a late timer firing after an interrupt already
// aborted is a no-op and can't relabel the interrupt as a timeout.
const ABORT_TIMEOUT = Symbol('daytona-bash-exec-timeout')

// One Map<sessionId, ShellRecord>.
const shells = new Map()

class ShellRecord {
  constructor(sessionId) {
    this.id = sessionId
    // Each session reads the real /workspace but its writes are private and
    // in-memory: this is the per-isolate boundary. mountPoint == root so user
    // paths (e.g. /workspace/foo) line up with the real sandbox path.
    const overlay = new OverlayFs({ root: WORKSPACE_ROOT, mountPoint: WORKSPACE_ROOT })
    this.bash = new Bash({
      fs: overlay,
      cwd: overlay.getMountPoint(),
      executionLimits: EXECUTION_LIMITS,
      // network omitted → curl/html-to-markdown unavailable (no egress).
      // python/javascript omitted → those runtimes stay off.
    })
    // FIFO so concurrent exec frames for one session serialize (mirrors the
    // per-context queue in repl_host.js / the Go session queue).
    this.queue = Promise.resolve()
    this.abort = null
    this.disposed = false
  }

  async run(code, envs, timeoutMs) {
    const controller = new AbortController()
    this.abort = controller
    const ms = timeoutMs > 0 ? timeoutMs : DEFAULT_EXEC_TIMEOUT_MS
    const timer = setTimeout(() => controller.abort(ABORT_TIMEOUT), ms)
    try {
      return await this.bash.exec(String(code == null ? '' : code), {
        env: envs && typeof envs === 'object' ? envs : undefined,
        signal: controller.signal,
      })
    } catch (e) {
      // Reclassify a timeout-caused abort (identified by the abort REASON, not by
      // a flag a late timer could set) so callers can tell it apart from an
      // explicit interrupt, which the interrupt op signals separately.
      if (controller.signal.aborted && controller.signal.reason === ABORT_TIMEOUT) {
        const timeoutErr = new Error('bash execution timed out')
        timeoutErr.name = 'TimeoutError'
        throw timeoutErr
      }
      throw e
    } finally {
      clearTimeout(timer)
      if (this.abort === controller) this.abort = null
    }
  }

  interrupt() {
    if (this.abort) {
      try { this.abort.abort() } catch { /* already aborted */ }
    }
  }

  dispose() {
    this.disposed = true
    this.interrupt()
  }
}

function ensureShell(sessionId) {
  let rec = shells.get(sessionId)
  if (!rec) {
    rec = new ShellRecord(sessionId)
    shells.set(sessionId, rec)
  }
  return rec
}

// Streaming exec for the standalone bash isolate. Aggregated stdout/stderr are
// emitted as single chunks (just-bash returns the whole result), followed by a
// terminal completed control chunk. Errors become an error chunk + completed.
function streamingExec(rec, cmd) {
  rec.queue = rec.queue.then(async () => {
    // A delete that landed while this job was queued must cancel it: the session
    // is gone from `shells` and its Go-side listener is being torn down, so
    // running user code now would execute against a disposed session. Still emit
    // the terminal so any waiter on the Go side unblocks.
    if (rec.disposed) {
      emit({ sessionId: rec.id, type: 'control', text: 'completed' })
      return
    }
    try {
      const result = await rec.run(cmd.code, cmd.envs, Number(cmd.execTimeoutMs) || 0)
      if (result.stdout) emit({ sessionId: rec.id, type: 'stdout', text: result.stdout })
      if (result.stderr) emit({ sessionId: rec.id, type: 'stderr', text: result.stderr })
    } catch (e) {
      const aborted = e && (e.name === 'AbortError' || /abort/i.test(String(e.message || '')))
      if (e && e.name === 'TimeoutError') {
        // Per-call timeout (run() classifies it via the abort reason): a real
        // failure the caller must see, else it would look like a clean completion.
        emit({ sessionId: rec.id, type: 'error', name: 'TimeoutError', value: e.message || 'bash execution timed out' })
      } else if (!aborted) {
        emit({ sessionId: rec.id, type: 'error', name: e && e.name ? e.name : 'BashError', value: e && e.message ? e.message : String(e) })
      }
      // else: explicit interrupt — the interrupt op handler already emitted the
      // single `interrupted` control frame; fall through to the one `completed`.
    }
    emit({ sessionId: rec.id, type: 'control', text: 'completed' })
  })
  return rec.queue
}

// One-shot reply-correlated call for the bash() bridges (TS/Python). Returns the
// aggregated result on a single control chunk so the Go side can resolve a
// pending request channel.
function bashCall(cmd) {
  const reply = cmd.reply || ''
  const rec = ensureShell(cmd.sessionId)
  rec.queue = rec.queue.then(async () => {
    // If a delete disposed this session while the call was queued, don't run user
    // code; resolve the bridge caller with an error so it doesn't block until its
    // timeout.
    if (rec.disposed) {
      emit({ type: 'control', text: 'bash-call-error', reply, value: 'session disposed' })
      return
    }
    try {
      const result = await rec.run(cmd.code, cmd.envs, Number(cmd.execTimeoutMs) || 0)
      emit({
        type: 'control',
        text: 'bash-call-result',
        reply,
        stdout: result.stdout || '',
        stderr: result.stderr || '',
        exitCode: typeof result.exitCode === 'number' ? result.exitCode : 0,
      })
    } catch (e) {
      emit({ type: 'control', text: 'bash-call-error', reply, value: e && e.message ? e.message : String(e) })
    }
  })
  return rec.queue
}

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
        if (shells.has(cmd.sessionId)) {
          emit({ sessionId: cmd.sessionId, type: 'error', name: 'ContextExistsError', value: 'context already exists' })
          return
        }
        ensureShell(cmd.sessionId)
        emit({ sessionId: cmd.sessionId, type: 'control', text: 'created' })
        break
      }
      case 'exec': {
        const rec = shells.get(cmd.sessionId)
        if (!rec) {
          emit({ sessionId: cmd.sessionId, type: 'error', name: 'ContextNotFoundError', value: 'context not found' })
          emit({ sessionId: cmd.sessionId, type: 'control', text: 'completed' })
          return
        }
        streamingExec(rec, cmd)
        break
      }
      case 'bash-call': {
        bashCall(cmd)
        break
      }
      case 'interrupt': {
        const rec = shells.get(cmd.sessionId)
        if (!rec) return
        rec.interrupt()
        emit({ sessionId: cmd.sessionId, type: 'control', text: 'interrupted' })
        break
      }
      case 'delete': {
        const rec = shells.get(cmd.sessionId)
        if (!rec) return
        rec.dispose()
        shells.delete(cmd.sessionId)
        emit({ sessionId: cmd.sessionId, type: 'control', text: 'deleted' })
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
  for (const rec of shells.values()) {
    try { rec.dispose() } catch { /* ignore */ }
  }
  process.exit(0)
})

emit({ type: 'control', text: 'host-ready' })
