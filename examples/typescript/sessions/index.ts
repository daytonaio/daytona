import { Daytona, SessionExpiredError, SessionInvalidatedError, SessionRunResult } from '@daytona/sdk'

// Daytona Sessions: zero-setup, sub-second code execution
// =======================================================
// Unlike `sandbox.process.codeRun`, sessions do NOT require you to provision
// a sandbox first. The Daytona API auto-allocates an ephemeral execution
// context per call (or a persistent one if you ask for it). Each example
// below highlights one product capability.

async function oneShotPython(daytona: Daytona) {
  // The simplest possible call: no sandbox, no context, no setup.
  console.log('=== one-shot Python ===')
  const result = await daytona.session.run('print(2 ** 10)', { language: 'python' })
  console.log('stdout:', result.stdout.trimEnd())
  console.log('took:', result.durationMs, 'ms')
}

async function oneShotTypeScriptWithImportsAndTla(daytona: Daytona) {
  // Showcases the headline TypeScript capabilities of the V8-session runtime:
  //   - real ESM `import` from a curated, pre-installed registry (zod here)
  //   - top-level await, including over network calls
  //   - the trailing expression's value comes back as a `display` payload
  console.log('\n=== one-shot TypeScript with ESM imports + top-level await ===')
  const code = `
import { z } from "zod"
const schema = z.object({ name: z.string(), age: z.number() })
const parsed = schema.safeParse({ name: "Ada", age: 36 })
console.log("parsed.success:", parsed.success)
console.log("data:", JSON.stringify(parsed.data))

const r = await fetch("https://httpbin.org/get?from=session")
const j = await r.json()
console.log("fetched OK, status:", r.status, "from:", j.args.from);

// Trailing expression — surfaced as a 'display' frame on the result.
({ ok: r.ok, status: r.status })
`.trim()
  const result = await daytona.session.run(code, { language: 'typescript' })
  process.stdout.write(result.stdout)
  if (result.error) {
    console.error('error:', result.error.name, result.error.value)
    return
  }
  for (const display of result.displays) {
    console.log('display formats:', display.formats.join(', '))
    if (display.data['text/plain']) console.log('display value:', display.data['text/plain'])
  }
}

async function statefulContext(daytona: Daytona) {
  // Persistent contexts keep variables alive across runs — perfect for
  // notebook-style workflows or interactive REPL frontends.
  console.log('\n=== stateful Python context (state survives across runs) ===')
  const ctx = await daytona.session.createSession({ language: 'python' })
  try {
    await daytona.session.run('counter = 0', { context: ctx })
    for (let i = 0; i < 3; i++) {
      const r = await daytona.session.run('counter += 1\nprint(f"counter = {counter}")', { context: ctx })
      process.stdout.write(r.stdout)
    }
  } finally {
    await daytona.session.deleteSession(ctx.id)
  }
}

async function multiContext(daytona: Daytona) {
  // Multiple persistent contexts can coexist on the same session instance.
  // Each one keeps its own globals — variables defined in `ctxA` are
  // invisible to `ctxB`, even within the same Python language. You can also
  // mix languages; here we run a Python pair alongside a TypeScript context
  // and list them all via listSessions() to confirm they're live.
  console.log('\n=== multiple coexisting contexts (independent state per context) ===')
  const ctxA = await daytona.session.createSession({ language: 'python' })
  const ctxB = await daytona.session.createSession({ language: 'python' })
  const ctxTs = await daytona.session.createSession({ language: 'typescript' })
  try {
    await daytona.session.run('identity = "alpha"', { context: ctxA })
    await daytona.session.run('identity = "beta"', { context: ctxB })
    // TypeScript: each `run()` compiles a fresh ES module, so `const`/`let`
    // declarations don't leak between runs. Use `globalThis` (or `var`) to
    // persist values across runs in the same context.
    await daytona.session.run('globalThis.identity = "gamma"', { context: ctxTs })

    const a = await daytona.session.run('print(f"A says: {identity}")', { context: ctxA })
    const b = await daytona.session.run('print(f"B says: {identity}")', { context: ctxB })
    const ts = await daytona.session.run('console.log(`TS says: ${globalThis.identity}`)', { context: ctxTs })
    process.stdout.write(a.stdout)
    process.stdout.write(b.stdout)
    process.stdout.write(ts.stdout)

    // listSessions() returns every persistent context in the org. Filter to
    // the three we just created so the example is deterministic when other
    // contexts happen to be live.
    const ours = new Set([ctxA.id, ctxB.id, ctxTs.id])
    const live = (await daytona.session.listSessions()).filter((c) => ours.has(c.id))
    const langs = live.map((c) => c.language).sort()
    console.log(`live contexts: ${live.length} (${langs.join(', ')})`)
  } finally {
    for (const ctx of [ctxA, ctxB, ctxTs]) {
      await daytona.session.deleteSession(ctx.id)
    }
  }
}

async function streamingWithHandlers(daytona: Daytona) {
  // runStream() opens a WebSocket and invokes the per-frame handlers as
  // each chunk arrives. The returned aggregated result is the same shape as
  // run(), so callers can use one or the other interchangeably.
  console.log('\n=== streaming Python with per-frame handlers ===')
  const code = `
for i in range(5):
    print(f"line {i}", flush=True)
import pandas as pd
pd.DataFrame({"x": [1, 2, 3], "y": [10, 20, 30]})
`.trim()
  const result: SessionRunResult = await daytona.session.runStream(code, {
    language: 'python',
    onStdout: (chunk) => process.stdout.write(`[stream] ${chunk}`),
    onError: (err) => console.error(`[stream-error] ${err.name}: ${err.value ?? ''}`),
    onDisplay: (d) => console.log(`[display] formats=${d.formats.join(',')}`),
  })
  console.log(
    `(${result.stdout.length} stdout chars, ${result.displays.length} display frames in ${result.durationMs}ms)`,
  )
}

async function errorSurfacing(daytona: Daytona) {
  // User errors — including TypeScript imports of unsupported Node builtins
  // — come back as a structured `error` frame, never as a thrown exception
  // on the SDK side.
  console.log('\n=== structured errors from user code ===')
  const py = await daytona.session.run('1 / 0', { language: 'python' })
  console.log(`Python: ${py.error?.name}: ${py.error?.value ?? ''}`)

  const ts = await daytona.session.run("import 'fs'", { language: 'typescript' })
  console.log(`TypeScript: ${ts.error?.name}: ${ts.error?.value ?? ''}`)
}

async function main() {
  // The Daytona client reads DAYTONA_API_URL / DAYTONA_API_KEY from env (or
  // pass them via the `new Daytona({...})` config object). No sandbox to
  // create, no image to build — the API server handles session provisioning
  // internally.
  const daytona = new Daytona()

  try {
    await oneShotPython(daytona)
    await oneShotTypeScriptWithImportsAndTla(daytona)
    await statefulContext(daytona)
    await multiContext(daytona)
    await streamingWithHandlers(daytona)
    await errorSurfacing(daytona)
  } catch (err) {
    if (err instanceof SessionInvalidatedError) {
      // The underlying sandbox was rolled (death / autostop). Drop the
      // context ref and re-create on next call.
      console.error('context was invalidated:', err.message)
    } else if (err instanceof SessionExpiredError) {
      // Hit the idle / absolute TTL. `err.reason` is 'idle' | 'absolute'.
      console.error('context expired:', err.reason, err.message)
    } else {
      throw err
    }
  }
}

main()
