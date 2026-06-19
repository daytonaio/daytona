/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Live test for sandbox recovery (the "container not found" bug).
 *   A. Stopped sandbox  -> next tool call auto-starts it and succeeds.
 *   B. Deleted sandbox  -> tool call throws a clear SandboxUnavailableError
 *                          (NOT a raw Docker error, and NOT run on the host).
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
const sandbox = await daytona.create({ labels: { 'created-by': 'pi-daytona-test' }, autoDeleteInterval: 60 })
const ops = createBashOps(sandbox)
const runOps = async (command) => {
  let out = ''
  const { exitCode } = await ops.exec(command, '/home/daytona', { onData: (b) => (out += b.toString()) })
  return { out: out.trim(), exitCode }
}

try {
  console.log('A. stopped sandbox auto-recovers')
  check((await runOps('echo before')).out === 'before', 'works while running')
  await sandbox.stop()
  await sandbox.refreshData()
  check(sandbox.state !== 'started', `sandbox stopped (state=${sandbox.state})`)
  const recovered = await runOps('echo recovered') // should auto-start + retry
  check(recovered.out === 'recovered', 'tool call auto-started the sandbox and succeeded', JSON.stringify(recovered))
  await sandbox.refreshData()
  check(sandbox.state === 'started', 'sandbox is running again after recovery')

  console.log('B. deleted sandbox -> clear error, no host fallback')
  await sandbox.delete()
  await new Promise((r) => setTimeout(r, 1500))
  let err
  try {
    await runOps('pwd')
  } catch (e) {
    err = e
  }
  // Check by name, not instanceof: jiti may load src/sandbox.ts as a distinct
  // module instance here vs. inside ops.ts, so the class objects differ.
  check(
    err && err.name === 'SandboxUnavailableError',
    'throws SandboxUnavailableError (not a raw Docker error)',
    err && err.name,
  )
  check(
    err && !/no such container|inspect sandbox/i.test(err.message),
    'error message is human-friendly',
    err && err.message?.slice(0, 80),
  )
  check(err && /restart pi/i.test(err.message), 'error tells the user how to recover')
} finally {
  // Note: only this run's sandbox — no sweep of all test sandboxes (would hit
  // concurrent runs); autoDeleteInterval is the backstop.
  await sandbox.delete().catch(() => {})
  console.log('  (cleaned up)')
}

console.log(`\n${fail === 0 ? 'PASS' : 'FAIL'}: ${pass} passed, ${fail} failed`)
if (fail) {
  console.log('Failures:\n' + failures.map((f) => `  - ${f}`).join('\n'))
  process.exitCode = 1
}
