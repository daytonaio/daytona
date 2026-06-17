/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Live test for the bash background-safety fix (createBashOps).
 * Verifies a backgrounded server does NOT hang the call, the server stays up,
 * and normal commands keep correct output + exit codes.
 *
 * Requires DAYTONA_API_KEY.
 */
import { createRequire } from 'node:module'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')
// Resolve the host package's exported entry with the ESM-aware resolver (the
// package ships an import-only `exports` map, so CJS require.resolve throws
// ERR_PACKAGE_PATH_NOT_EXPORTED); it walks hoisting too. Then anchor a CJS
// require there for jiti (which is CJS-compatible).
const hostEntry = fileURLToPath(import.meta.resolve('@earendil-works/pi-coding-agent'))
const { createJiti } = createRequire(hostEntry)('jiti')
const jiti = createJiti(import.meta.url, {
  moduleCache: false,
  alias: { '@earendil-works/pi-coding-agent': hostEntry },
})
const { Daytona } = await import('@daytona/sdk')

let pass = 0,
  fail = 0
const failures = []
const check = (cond, msg, detail) => {
  if (cond) {
    pass++
    console.log(`  ✓ ${msg}`)
  } else {
    fail++
    failures.push(msg)
    console.log(`  ✗ ${msg}${detail ? ` — ${detail}` : ''}`)
  }
}

const { createBashOps } = await jiti.import(path.join(root, 'src/ops.ts'))

const daytona = new Daytona()
const sandbox = await daytona.create({ ephemeral: true, labels: { 'created-by': 'pi-daytona-test' } })
try {
  const home = (await sandbox.getUserHomeDir()) ?? '/home/daytona'
  const ops = createBashOps(sandbox)
  // Bound every exec with a watchdog: the whole point of this test is that a
  // backgrounded command must NOT hang ops.exec. Without a timeout a regression
  // would stall forever (CI hang) instead of failing — so race against a timer
  // and reject if exec ever blocks past the limit.
  const runOps = async (command, timeoutMs = 30000) => {
    let out = ''
    const started = Date.now()
    let timer
    const watchdog = new Promise((_, reject) => {
      timer = setTimeout(() => reject(new Error(`ops.exec timed out after ${timeoutMs}ms: ${command}`)), timeoutMs)
    })
    try {
      const { exitCode } = await Promise.race([
        ops.exec(command, home, { onData: (b) => (out += b.toString()) }),
        watchdog,
      ])
      return { out: out.trim(), exitCode, ms: Date.now() - started }
    } finally {
      clearTimeout(timer)
    }
  }
  // Poll instead of a fixed sleep: server startup time varies by sandbox, so
  // retry the curl until it answers 200 (up to ~9s) rather than guessing.
  const waitForHttp = async (url) => {
    let last
    for (let i = 0; i < 30; i++) {
      last = await runOps(`curl -s -o /dev/null -w '%{http_code}' ${url} || echo FAIL`)
      if (/200/.test(last.out)) return last
      await new Promise((r) => setTimeout(r, 300))
    }
    return last
  }

  console.log('normal commands')
  const echo = await runOps('echo hello')
  check(echo.out === 'hello' && echo.exitCode === 0, 'echo: output + exit 0', JSON.stringify(echo))
  const f = await runOps('false')
  check(f.exitCode === 1, 'false: exit code 1 preserved', JSON.stringify(f))
  const err = await runOps('echo to-stderr >&2')
  check(/to-stderr/.test(err.out) && err.exitCode === 0, 'stderr captured + merged', JSON.stringify(err))
  const ex = await runOps('echo before; exit 7')
  check(
    ex.exitCode === 7 && /before/.test(ex.out),
    'exit N inside command: code + prior output kept',
    JSON.stringify(ex),
  )
  const missing = await runOps('ls /no/such/path')
  check(missing.exitCode !== 0 && missing.out.length > 0, 'ls missing: nonzero + error text', JSON.stringify(missing))

  console.log('backgrounded server (the reported hang)')
  const bg = await runOps('python3 -m http.server 8080 &')
  check(bg.ms < 8000, `did NOT hang (returned in ${bg.ms}ms)`, JSON.stringify(bg))
  // Server should be alive: curl it from inside the sandbox.
  const curl = await waitForHttp('http://localhost:8080/')
  check(/200/.test(curl.out), 'backgrounded server is alive (HTTP 200)', JSON.stringify(curl))

  console.log('background + foreground in one command')
  const mix = await runOps('python3 -m http.server 8081 & echo started')
  check(mix.ms < 8000 && /started/.test(mix.out), `mixed bg+fg returns fast with fg output`, JSON.stringify(mix))
  const curl2 = await waitForHttp('http://localhost:8081/')
  check(/200/.test(curl2.out), 'second server alive (HTTP 200)', JSON.stringify(curl2))
} catch (e) {
  // A thrown error (e.g. a runOps watchdog timeout) would otherwise abort with a
  // raw rejection; record it as a failed check so the summary below still prints.
  check(false, 'aborted early', e?.message)
} finally {
  await sandbox.delete().catch(() => {})
  console.log('  (sandbox deleted)')
}

console.log(`\n${fail === 0 ? 'PASS' : 'FAIL'}: ${pass} passed, ${fail} failed`)
if (fail) {
  console.log('Failures:\n' + failures.map((f) => `  - ${f}`).join('\n'))
  process.exitCode = 1
}
