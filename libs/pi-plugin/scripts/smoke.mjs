/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Smoke test: load the extension exactly as Pi does (via jiti) against a stub
 * ExtensionAPI, and assert it registers the expected flags, tools, and events
 * without throwing. Does NOT require a Daytona API key or network.
 */

import { createRequire } from 'node:module'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')

// Use the same jiti loader Pi bundles. Resolve the host package's exported entry
// with the ESM-aware resolver (the package ships an import-only `exports` map, so
// CJS require.resolve throws ERR_PACKAGE_PATH_NOT_EXPORTED); it walks hoisting too.
// Then anchor a CJS require there for jiti (which is CJS-compatible).
const hostEntry = fileURLToPath(import.meta.resolve('@earendil-works/pi-coding-agent'))
const { createJiti } = createRequire(hostEntry)('jiti')
const jiti = createJiti(import.meta.url)

const flags = []
const tools = []
const events = []
const commands = []

const stubPi = {
  registerFlag: (name) => flags.push(name),
  registerTool: (tool) => tools.push(tool?.name),
  registerCommand: (name) => commands.push(name),
  on: (event) => events.push(event),
  getFlag: () => undefined,
}

const mod = await jiti.import(path.join(root, 'index.ts'))
const factory = mod.default ?? mod

if (typeof factory !== 'function') {
  throw new Error('Extension default export is not a function')
}

// Await like Pi does: the factory is synchronous today, but Pi awaits the
// result, so this stays faithful (and robust if it ever becomes async).
await factory(stubPi)

const expectedFlags = ['daytona', 'repo', 'branch', 'snapshot', 'public']
const expectedTools = ['bash', 'read', 'write', 'edit', 'ls', 'find', 'grep', 'preview_url']
const expectedEvents = ['user_bash', 'session_start', 'before_agent_start', 'agent_end', 'session_shutdown']
const expectedCommands = ['sandbox', 'merge', 'pr', 'compare', 'github']

function assertContains(label, actual, expected) {
  const missing = expected.filter((e) => !actual.includes(e))
  if (missing.length) {
    throw new Error(`${label}: missing ${JSON.stringify(missing)} (got ${JSON.stringify(actual)})`)
  }
  console.log(`✓ ${label}: ${expected.join(', ')}`)
}

assertContains('flags', flags, expectedFlags)
assertContains('tools', tools, expectedTools)
assertContains('events', events, expectedEvents)
assertContains('commands', commands, expectedCommands)

console.log('\nSmoke test passed: extension loads and registers cleanly.')
