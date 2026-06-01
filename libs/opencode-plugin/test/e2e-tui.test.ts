/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 *
 * End-to-end test that drives the *real* OpenCode TUI (via tmux, the same way
 * you would by hand) against the latest released opencode build:
 *
 *   open opencode → /warp → select "Daytona" → wait for the sandbox → send a message
 *
 * Hard assertions (the plugin's responsibility):
 *   - the workspace provisions: a Daytona sandbox is created and its remote
 *     opencode server becomes healthy (create() completes, no hang)
 *   - the remote actually answers: the sandbox's session contains a non-empty
 *     assistant reply containing a magic token we asked the model to echo
 *
 * Informational (NOT asserted): whether the reply renders in the host TUI.
 * At time of writing it does not — OpenCode's experimental-workspace global
 * sync delivers session state counts but not message content, so the TUI stays
 * blank even though the remote produced the answer. That gap is upstream, not
 * in this plugin, so we only log it rather than fail on it.
 *
 * Requires: DAYTONA_API_KEY, tmux, and a released `opencode` binary
 * (OPENCODE_BIN or ~/.opencode/bin/opencode). Skipped otherwise.
 */

import { afterAll, beforeAll, describe, expect, test } from 'bun:test'
import { spawnSync } from 'node:child_process'
import { mkdtemp, writeFile, rm, symlink, readFile } from 'node:fs/promises'
import { tmpdir, homedir } from 'node:os'
import { join, resolve } from 'node:path'
import { Daytona } from '@daytona/sdk'

const HAS_KEY = Boolean(process.env.DAYTONA_API_KEY)
const HAS_TMUX = spawnSync('tmux', ['-V'], { encoding: 'utf8' }).status === 0
const OPENCODE_BIN = process.env.OPENCODE_BIN || join(homedir(), '.opencode/bin/opencode')
const HAS_BIN = spawnSync(OPENCODE_BIN, ['--version'], { encoding: 'utf8' }).status === 0
const ENABLED = HAS_KEY && HAS_TMUX && HAS_BIN

// tmux server isolated on its own socket so we don't touch the user's sessions.
const SOCKET = 'daytona-e2e'
const WINDOW = 'oc'
const PLUGIN_LOG = '/tmp/daytona-plugin.log'
// Distinctive token the model is asked to echo — won't collide with TUI chrome.
const MAGIC = 'PINEAPPLE7391'
// .opencode lives one level up from test/.
const PLUGIN_DOT_OPENCODE = resolve(import.meta.dir, '..', '.opencode')

function tmux(...args: string[]) {
  return spawnSync('tmux', ['-L', SOCKET, ...args], { encoding: 'utf8' })
}
function capture(): string {
  return tmux('capture-pane', '-t', WINDOW, '-p').stdout || ''
}
function sendKeys(...keys: string[]) {
  tmux('send-keys', '-t', WINDOW, ...keys)
}
function sleep(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms))
}
async function readLog(): Promise<string> {
  return readFile(PLUGIN_LOG, 'utf8').catch(() => '')
}
async function waitForScreen(pred: (text: string) => boolean, timeoutMs: number, label: string): Promise<string> {
  const start = Date.now()
  let last = ''
  while (Date.now() - start < timeoutMs) {
    last = capture()
    if (pred(last)) return last
    await sleep(1500)
  }
  throw new Error(`timed out (${timeoutMs}ms) waiting for: ${label}\n--- last screen ---\n${last}`)
}

describe.skipIf(!ENABLED)('e2e tui (daytona workspace)', () => {
  let projectDir: string
  let daytona: Daytona
  let createdSandboxName: string | null = null

  beforeAll(async () => {
    daytona = new Daytona({ apiKey: process.env.DAYTONA_API_KEY })

    // 1. Throwaway project with the plugin symlinked in (loads TS source directly).
    projectDir = await mkdtemp(join(tmpdir(), 'daytona-e2e-'))
    await symlink(PLUGIN_DOT_OPENCODE, join(projectDir, '.opencode'))
    await writeFile(join(projectDir, 'README.md'), '# e2e test project\n')
    await writeFile(join(projectDir, '.env'), `DAYTONA_API_KEY=${process.env.DAYTONA_API_KEY}\n`)
    const git = (...a: string[]) => spawnSync('git', a, { cwd: projectDir })
    git('init', '-q')
    git('-c', 'user.email=e2e@test.dev', '-c', 'user.name=e2e', 'add', '-A')
    git('-c', 'user.email=e2e@test.dev', '-c', 'user.name=e2e', 'commit', '-qm', 'init')

    // 2. Launch the real TUI in tmux.
    tmux('kill-server')
    tmux('new-session', '-d', '-s', WINDOW, '-x', '200', '-y', '50')
    sendKeys(
      `cd ${projectDir} && OPENCODE_EXPERIMENTAL_WORKSPACES=true DAYTONA_API_KEY=${process.env.DAYTONA_API_KEY} ${OPENCODE_BIN}`,
      'Enter',
    )
    await waitForScreen((t) => /Ask anything|Build ·/.test(t), 40_000, 'TUI home screen')
  }, 90_000)

  afterAll(async () => {
    tmux('kill-server')
    if (createdSandboxName) {
      const sb = await daytona.get(createdSandboxName).catch(() => undefined)
      if (sb) await daytona.delete(sb).catch(() => undefined)
    }
    if (projectDir) await rm(projectDir, { recursive: true, force: true }).catch(() => undefined)
  })

  test('creates a Daytona workspace and the remote answers a chat message', async () => {
    // --- /warp → select Daytona (Worktree is first, Daytona second) ---
    const logBefore = (await readLog()).length
    sendKeys('/warp')
    await sleep(800)
    sendKeys('Enter')
    await waitForScreen((t) => /Daytona\s+Create a remote/.test(t), 10_000, 'warp picker')
    sendKeys('Down')
    await sleep(800)
    sendKeys('Enter')

    // --- wait for create() to finish (watch the plugin's own debug log) ---
    let newLog = ''
    const start = Date.now()
    while (Date.now() - start < 180_000) {
      newLog = (await readLog()).slice(logBefore)
      if (/create: server healthy; done/.test(newLog)) break
      if (/create: ERROR/.test(newLog)) throw new Error('create() errored:\n' + newLog)
      await sleep(2000)
    }
    expect(newLog, 'create() should reach "server healthy; done"').toMatch(/create: server healthy; done/)

    const name = newLog.match(/create: start name=(\S+)/)?.[1]
    const id = newLog.match(/d\.create\(\) returned sandbox id=(\S+)/)?.[1]
    expect(name, 'workspace name in log').toBeTruthy()
    expect(id, 'sandbox id in log').toBeTruthy()
    if (!name || !id) throw new Error('could not parse workspace name/id from plugin log')
    createdSandboxName = `opencode-${name}`
    console.log(`[e2e] created workspace=${name} sandbox=${createdSandboxName} id=${id}`)

    // --- HARD ASSERT: the sandbox exists and its remote server is healthy ---
    const sandbox = await daytona.get(createdSandboxName)
    expect(sandbox.state).toBe('started')
    const health = await sandbox.process.executeCommand('curl -fsS http://127.0.0.1:3096/global/health')
    expect(health.exitCode, 'remote /global/health').toBe(0)
    expect(health.result).toContain('"healthy":true')

    // --- send a chat message asking the model to echo a magic token ---
    await waitForScreen((t) => /Workspace\s/.test(t) || t.includes(name), 20_000, 'TUI entered workspace')
    sendKeys(`Reply with exactly the word ${MAGIC} and nothing else.`)
    await sleep(800)
    sendKeys('Enter')

    // --- HARD ASSERT: the remote generates a reply containing the token ---
    let assistantText = ''
    const t2 = Date.now()
    while (Date.now() - t2 < 60_000) {
      const sres = await sandbox.process.executeCommand('curl -s http://127.0.0.1:3096/session')
      const sessions = JSON.parse(sres.result || '[]')
      if (sessions.length) {
        sessions.sort((a: any, b: any) => (b.time?.updated || 0) - (a.time?.updated || 0))
        const mres = await sandbox.process.executeCommand(
          `curl -s http://127.0.0.1:3096/session/${sessions[0].id}/message`,
        )
        const msgs = JSON.parse(mres.result || '[]')
        assistantText = msgs
          .filter((m: any) => m.info?.role === 'assistant')
          .flatMap((m: any) => (m.parts || []).filter((p: any) => p.type === 'text').map((p: any) => p.text))
          .join(' ')
        if (assistantText.includes(MAGIC)) break
      }
      await sleep(3000)
    }
    expect(assistantText, 'remote assistant reply should contain the magic token').toContain(MAGIC)
    console.log(`[e2e] remote reply contains "${MAGIC}" ✅ — remote generated the answer`)

    // --- INFORMATIONAL: did the host TUI render the reply? (upstream sync gap) ---
    await sleep(4000)
    const tui = capture()
    const renderedInTui = tui.includes(MAGIC)
    if (renderedInTui) {
      console.log('[e2e] TUI rendered the reply — host<->remote sync delivered message content.')
    } else {
      console.log(
        '[e2e] NOTE: remote produced the reply but the host TUI did NOT render it — ' +
          'reproduces the upstream OpenCode global-sync gap (state counts sync, message content does not). ' +
          'Not asserted; outside this plugin.',
      )
    }
  }, 300_000)
})
