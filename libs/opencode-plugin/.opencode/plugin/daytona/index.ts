/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * OpenCode plugin that registers Daytona sandboxes as a workspace adaptor.
 * Each session spawns a remote sandbox running `opencode serve`; tool calls
 * are proxied over the preview URL rather than invoked locally.
 */

import { spawn as nodeSpawn } from 'node:child_process'
import { appendFileSync } from 'node:fs'
import { mkdir, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { Daytona } from '@daytona/sdk'
import type { PluginInput, WorkspaceAdapter } from '@opencode-ai/plugin'

import { buildSandboxInstructions } from './instructions'

// Lazy so DAYTONA_API_KEY is read at use-time, not module-load time.
let daytonaClient: Daytona | undefined

// Accessor for the shared Daytona client; instantiates it on first call.
function getDaytona(): Daytona {
  if (daytonaClient == null) {
    daytonaClient = new Daytona({
      apiKey: process.env.DAYTONA_API_KEY,
    })
  }
  return daytonaClient
}

// Cache preview links so we don't re-check the sandbox on every target() call
// (opencode calls target() in bursts). verifiedAt bounds how long we trust the
// entry before re-confirming the sandbox is up.
type PreviewEntry = { url: string; token: string; verifiedAt: number }
const previewCache = new Map<string, PreviewEntry>()
const PREVIEW_TTL_MS = 15_000

// Namespace sandboxes to distinguish them from non-opencode sandboxes
// in the same Daytona account.
function sandboxName(name: string): string {
  return `opencode-${name}`
}

const REPO_PATH = '/home/daytona/workspace/repo'
const ROOT_PATH = '/home/daytona/workspace'
const LOCAL_BIN = '/home/daytona/opencode'
const INSTALL_BIN = '/home/daytona/.opencode/bin/opencode'
const SERVER_PORT = 3096
const HEALTH_URL = `http://127.0.0.1:${SERVER_PORT}/global/health`

// Pin the opencode version installed in the sandbox. Passing VERSION to the
// installer skips its "latest release" lookup against api.github.com, which is
// rate-limited (HTTP 429) and fails with "Failed to fetch version information"
// when many sandboxes install in a short window. Bump as needed.
const OPENCODE_VERSION = '1.15.13'

// POSIX-safe single-quote escape: close quote, emit literal ', reopen quote.
function sh(value: string): string {
  return `'${value.replace(/'/g, "'\"'\"'")}'`
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

// Reject if `promise` doesn't settle within `ms`. The sandbox stops servicing
// new commands while opencode runs its one-time DB migration on first start, so
// a poll issued during that window can hang indefinitely. Without a client-side
// timeout that single stuck call wedges the health-poll loop — and thus the
// whole workspace creation — forever, even after the server is healthy.
function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  return Promise.race([
    promise,
    new Promise<T>((_, reject) => setTimeout(() => reject(new Error(`timed out after ${ms}ms`)), ms)),
  ])
}

// Debug log to a fixed file (stdout is kept clean for the UI). Synchronous append
// so entries flush even if the very next step hangs. Tail with:
//   tail -f /tmp/daytona-plugin.log
const LOG_FILE = '/tmp/daytona-plugin.log'
function debug(msg: string): void {
  try {
    appendFileSync(LOG_FILE, `${new Date().toISOString()} [daytona] ${msg}\n`)
  } catch {
    // never let logging break the plugin
  }
}

type SandboxHandle = Awaited<ReturnType<Daytona['get']>>

// Build the workspace target opencode connects to (tool calls + the global-sync
// /global/event stream both go here).
function toTarget(link: { url: string; token: string }) {
  return {
    type: 'remote' as const,
    url: link.url,
    headers: {
      'x-daytona-preview-token': link.token,
      'x-daytona-skip-preview-warning': 'true',
      'x-opencode-directory': REPO_PATH,
    },
  }
}

// Command that (re)starts `opencode serve` in the sandbox. Prefers a pre-baked
// binary if the snapshot ships one; otherwise the version installed at create.
function serverLaunchCmd(): string {
  return `cd ${sh(REPO_PATH)} && exe=${sh(LOCAL_BIN)} && if [ ! -x "$exe" ]; then exe=${sh(INSTALL_BIN)}; fi && nohup env "$exe" serve --hostname 0.0.0.0 --port ${SERVER_PORT} >/tmp/opencode.log 2>&1 </dev/null &`
}

// True if the in-sandbox opencode server answers /global/health. Bounded by
// withTimeout so a command issued during the DB migration can't hang it.
async function isServerHealthy(sandbox: SandboxHandle): Promise<boolean> {
  try {
    const r = await withTimeout(sandbox.process.executeCommand(`curl -fsS ${sh(HEALTH_URL)}`), 5000)
    return r.exitCode === 0
  } catch {
    return false
  }
}

// Ensure `opencode serve` is up, launching it and waiting if not. Used by
// create() and by target() after resuming a timed-out sandbox — the server
// process does not survive a sandbox stop/start.
async function ensureServerRunning(sandbox: SandboxHandle): Promise<void> {
  if (await isServerHealthy(sandbox)) return
  debug(`ensureServer: (re)launching opencode serve on ${sandbox.id}`)
  await sandbox.process.executeCommand(serverLaunchCmd())
  for (let i = 0; i < 60; i++) {
    if (await isServerHealthy(sandbox)) {
      debug('ensureServer: healthy')
      return
    }
    await sleep(1000)
  }
  const log = await withTimeout(
    sandbox.process.executeCommand('test -f /tmp/opencode.log && cat /tmp/opencode.log || true'),
    5000,
  ).catch(() => undefined)
  throw new Error(log?.result || 'Daytona workspace server did not become ready in time')
}

// Runs a host-side command (not in the sandbox); rejects with aggregated stderr on non-zero exit.
async function spawnAsync(cmd: string[], options: { cwd?: string; env?: NodeJS.ProcessEnv } = {}): Promise<void> {
  return new Promise((resolve, reject) => {
    const proc = nodeSpawn(cmd[0], cmd.slice(1), {
      cwd: options.cwd,
      env: options.env,
      stdio: ['ignore', 'pipe', 'pipe'],
    })

    let stderr = ''
    proc.stderr?.on('data', (data: Buffer) => {
      stderr += data.toString()
    })

    proc.on('close', (code: number | null) => {
      if (code === 0) {
        resolve()
      } else {
        reject(new Error(stderr || `Command failed with exit code ${code}: ${cmd.join(' ')}`))
      }
    })

    proc.on('error', reject)
  })
}

// Drop undefined values so Daytona (whose envVars wants Record<string, string>) accepts the map.
function toEnvVars(env: Record<string, string | undefined>): Record<string, string> {
  const result: Record<string, string> = {}
  for (const [key, value] of Object.entries(env)) {
    if (value !== undefined) result[key] = value
  }
  return result
}

export const DaytonaWorkspacePlugin = async (input: PluginInput) => {
  const { experimental_workspace, worktree, project } = input

  if (!process.env.DAYTONA_API_KEY) {
    console.warn('[daytona] DAYTONA_API_KEY is not set - Daytona workspaces will not work')
  }

  const adaptor: WorkspaceAdapter = {
    name: 'Daytona',
    description: 'Create a remote Daytona sandbox workspace',

    // No-op: opencode's default config is fine as-is.
    configure(config) {
      return config
    },

    // Provision a fresh sandbox: upload the repo, install opencode, start `opencode serve`.
    async create(config, env) {
      debug(`create: start name=${config.name} branch=${config.branch ?? '(default)'} worktree=${worktree}`)
      if (!process.env.DAYTONA_API_KEY) {
        throw new Error('DAYTONA_API_KEY environment variable is not set')
      }

      const temp = join(tmpdir(), `opencode-daytona-${Date.now()}`)
      const d = getDaytona()
      debug(
        `create: calling d.create() sandbox=${sandboxName(config.name)} envKeys=${Object.keys(toEnvVars(env)).join(',')}`,
      )
      const sandbox = await d.create({
        name: sandboxName(config.name),
        envVars: toEnvVars(env),
        // Never auto-delete. A deleted sandbox orphans its OpenCode workspace
        // entry, which then floods the host with "failed to connect to global
        // sync". Idle sandboxes only auto-stop (pause); target() resumes them on
        // demand. (-1 disables auto-delete regardless of account default.)
        autoDeleteInterval: -1,
      })
      debug(`create: d.create() returned sandbox id=${sandbox.id} state=${sandbox.state}`)

      try {
        // Run a sandbox command, throwing its output on non-zero exit. Output is
        // surfaced only on failure: streaming every command's stdout (the opencode
        // installer's progress bar and banner, tar, etc.) floods the host terminal
        // with noise during workspace creation.
        const run = async (command: string): Promise<void> => {
          const label = command.length > 70 ? command.slice(0, 70) + '…' : command
          debug(`run: ⟶ ${label}`)
          const result = await sandbox.process.executeCommand(command)
          debug(`run: ⟵ exit=${result.exitCode} (${label})`)
          if (result.exitCode !== 0) {
            throw new Error(result.result || `Sandbox command failed: ${command}`)
          }
        }

        await mkdir(temp, { recursive: true })
        const dir = join(temp, 'repo')
        const tar = join(temp, 'repo.tgz')
        const source = `file://${worktree}`

        const cloneArgs = ['git', 'clone', '--depth', '1', '--no-local']
        if (config.branch) {
          cloneArgs.push('--branch', config.branch)
        }
        cloneArgs.push(source, dir)

        debug(`create: git clone ${source} -> ${dir}`)
        await spawnAsync(cloneArgs, { cwd: tmpdir() })
        debug('create: git clone done; building tarball')
        // Strip the host's .opencode/: it's host-side opencode config (agents,
        // plugins, instructions). The sandbox runs its own opencode and copying
        // the host's would clobber the plugin's own .opencode/ writes — and
        // breaks outright if .opencode is a symlink (the local-dev recipe).
        //
        // --no-xattrs and COPYFILE_DISABLE=1 keep macOS bsdtar from embedding
        // Apple extended attributes (e.g. com.apple.provenance) as LIBARCHIVE.xattr.*
        // pax headers and AppleDouble ._* entries. Without them the sandbox's GNU
        // tar floods stdout with "Ignoring unknown extended header keyword" warnings
        // and litters the repo with ._* files. Both are no-ops on Linux hosts.
        await spawnAsync(['tar', '--no-xattrs', '--exclude=repo/.opencode', '-czf', tar, '-C', temp, 'repo'], {
          env: { ...process.env, COPYFILE_DISABLE: '1' },
        })
        debug('create: tarball built; uploading repo.tgz')

        await sandbox.fs.uploadFile(tar, 'repo.tgz')
        debug('create: repo.tgz uploaded; extracting in sandbox')
        await run(
          `rm -rf ${sh(REPO_PATH)} && mkdir -p ${sh(ROOT_PATH)} && tar -xzf "$HOME/repo.tgz" -C "$HOME/workspace" && rm "$HOME/repo.tgz"`,
        )

        debug(`create: installing opencode ${OPENCODE_VERSION} in sandbox`)
        await run(
          `mkdir -p "$HOME/.opencode/bin" && curl -fsSL https://opencode.ai/install | VERSION=${OPENCODE_VERSION} OPENCODE_INSTALL_DIR="$HOME/.opencode/bin" bash`,
        )
        debug('create: opencode install done; uploading project id')

        await sandbox.fs.uploadFile(Buffer.from(`${project.id}\n`), `${REPO_PATH}/.git/opencode`)

        // Derive preview URL template from actual getPreviewLink response to avoid hardcoding the proxy domain.
        debug('create: fetching sample preview link (port 8080)')
        const samplePreview = await sandbox.getPreviewLink(8080)
        const previewUrlTemplate = samplePreview.url.replace(/^(https?:\/\/)\d+-/, '$1<port>-')
        debug(`create: preview template=${previewUrlTemplate}`)

        const instructions = buildSandboxInstructions({ repoPath: REPO_PATH, previewUrlTemplate })
        await sandbox.fs.uploadFile(Buffer.from(instructions), `${REPO_PATH}/.opencode/instructions/daytona.md`)
        debug('create: instructions uploaded')

        const opencodeConfig = JSON.stringify(
          {
            $schema: 'https://opencode.ai/config.json',
            instructions: ['.opencode/instructions/daytona.md'],
          },
          null,
          2,
        )
        await sandbox.fs.uploadFile(Buffer.from(opencodeConfig), `${REPO_PATH}/opencode.json`)
        debug('create: opencode.json uploaded; starting server')

        debug('create: starting opencode server')
        await ensureServerRunning(sandbox)
        debug('create: server healthy; done ✅')
        return
      } catch (err) {
        debug(`create: ERROR ${err instanceof Error ? err.message : String(err)}`)
        // Don't leak the sandbox if anything after Daytona.create() throws.
        await d.delete(sandbox).catch(() => undefined)
        throw err
      } finally {
        await rm(temp, { recursive: true, force: true }).catch(() => undefined)
      }
    },

    // Tear down the sandbox and drop its cached preview link.
    async remove(config) {
      debug(`remove: start name=${config.name}`)
      const d = getDaytona()
      const sandbox = await d.get(sandboxName(config.name)).catch(() => undefined)
      if (!sandbox) {
        debug(`remove: no sandbox found for ${sandboxName(config.name)}; nothing to do`)
        return
      }
      await d.delete(sandbox)
      previewCache.delete(config.name)
      debug(`remove: deleted sandbox id=${sandbox.id}`)
    },

    // Remote endpoint opencode proxies tool calls to — and where the workspace
    // global-sync connects (it dials <url>/global/event). A paused sandbox here
    // surfaces on the host as "failed to connect to global sync". So resume the
    // sandbox on access: a timed-out (auto-stopped) workspace isn't dead, just
    // paused, and we bring it back rather than letting opencode treat it as gone.
    async target(config) {
      // Fast path: recently-verified link (target() is called in bursts).
      const cached = previewCache.get(config.name)
      if (cached && Date.now() - cached.verifiedAt < PREVIEW_TTL_MS) {
        return toTarget(cached)
      }

      const sandbox = await getDaytona().get(sandboxName(config.name))
      debug(`target: name=${config.name} state=${sandbox.state}`)
      if (sandbox.state !== 'started') {
        debug(`target: resuming ${sandboxName(config.name)} (state=${sandbox.state})`)
        await sandbox.start()
      }
      // The serve process doesn't survive a stop/start, so (re)launch if needed.
      await ensureServerRunning(sandbox)

      const link = await sandbox.getPreviewLink(SERVER_PORT)
      const entry: PreviewEntry = { url: link.url, token: link.token, verifiedAt: Date.now() }
      previewCache.set(config.name, entry)
      debug(`target: returning url=${entry.url}`)
      return toTarget(entry)
    },
  }

  experimental_workspace.register('daytona', adaptor)
  debug(`plugin loaded; registered 'daytona' adaptor (log file: ${LOG_FILE})`)

  return {}
}

export default DaytonaWorkspacePlugin
