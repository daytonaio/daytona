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

// Cache preview links so we don't refetch on every target() call.
const previewCache = new Map<string, { url: string; token: string }>()

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

// POSIX-safe single-quote escape: close quote, emit literal ', reopen quote.
function sh(value: string): string {
  return `'${value.replace(/'/g, "'\"'\"'")}'`
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
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
      if (!process.env.DAYTONA_API_KEY) {
        throw new Error('DAYTONA_API_KEY environment variable is not set')
      }

      const temp = join(tmpdir(), `opencode-daytona-${Date.now()}`)
      const d = getDaytona()
      const sandbox = await d.create({
        name: sandboxName(config.name),
        envVars: toEnvVars(env),
      })

      try {
        // Stream sandbox command output to host stdout; throw on non-zero exit.
        const run = async (command: string): Promise<void> => {
          const result = await sandbox.process.executeCommand(command)
          if (result.result) {
            process.stdout.write(result.result)
          }
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

        await spawnAsync(cloneArgs, { cwd: tmpdir() })
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

        await sandbox.fs.uploadFile(tar, 'repo.tgz')
        await run(
          `rm -rf ${sh(REPO_PATH)} && mkdir -p ${sh(ROOT_PATH)} && tar -xzf "$HOME/repo.tgz" -C "$HOME/workspace" && rm "$HOME/repo.tgz"`,
        )

        await run(
          `mkdir -p "$HOME/.opencode/bin" && OPENCODE_INSTALL_DIR="$HOME/.opencode/bin" curl -fsSL https://opencode.ai/install | bash`,
        )

        await sandbox.fs.uploadFile(Buffer.from(`${project.id}\n`), `${REPO_PATH}/.git/opencode`)

        // Derive preview URL template from actual getPreviewLink response to avoid hardcoding the proxy domain.
        const samplePreview = await sandbox.getPreviewLink(8080)
        const previewUrlTemplate = samplePreview.url.replace(/^(https?:\/\/)\d+-/, '$1<port>-')

        const instructions = buildSandboxInstructions({ repoPath: REPO_PATH, previewUrlTemplate })
        await sandbox.fs.uploadFile(Buffer.from(instructions), `${REPO_PATH}/.opencode/instructions/daytona.md`)

        const opencodeConfig = JSON.stringify(
          {
            $schema: 'https://opencode.ai/config.json',
            instructions: ['.opencode/instructions/daytona.md'],
          },
          null,
          2,
        )
        await sandbox.fs.uploadFile(Buffer.from(opencodeConfig), `${REPO_PATH}/opencode.json`)

        // Prefer a pre-baked opencode binary if the snapshot ships one;
        // otherwise fall back to the version installed above.
        await run(
          `cd ${sh(REPO_PATH)} && exe=${sh(LOCAL_BIN)} && if [ ! -x "$exe" ]; then exe=${sh(INSTALL_BIN)}; fi && nohup env "$exe" serve --hostname 0.0.0.0 --port ${SERVER_PORT} >/tmp/opencode.log 2>&1 </dev/null &`,
        )

        for (let i = 0; i < 60; i++) {
          const result = await sandbox.process.executeCommand(`curl -fsS ${sh(HEALTH_URL)}`)
          if (result.exitCode === 0) {
            return
          }
          await sleep(1000)
        }

        const log = await sandbox.process.executeCommand('test -f /tmp/opencode.log && cat /tmp/opencode.log || true')
        throw new Error(log.result || 'Daytona workspace server did not become ready in time')
      } catch (err) {
        // Don't leak the sandbox if anything after Daytona.create() throws.
        await d.delete(sandbox).catch(() => undefined)
        throw err
      } finally {
        await rm(temp, { recursive: true, force: true }).catch(() => undefined)
      }
    },

    // Tear down the sandbox and drop its cached preview link.
    async remove(config) {
      const d = getDaytona()
      const sandbox = await d.get(sandboxName(config.name)).catch(() => undefined)
      if (!sandbox) return
      await d.delete(sandbox)
      previewCache.delete(config.name)
    },

    // Remote endpoint opencode proxies tool calls to.
    async target(config) {
      let link = previewCache.get(config.name)
      if (!link) {
        const sandbox = await getDaytona().get(sandboxName(config.name))
        link = await sandbox.getPreviewLink(SERVER_PORT)
        previewCache.set(config.name, link)
      }
      return {
        type: 'remote' as const,
        url: link.url,
        headers: {
          'x-daytona-preview-token': link.token,
          'x-daytona-skip-preview-warning': 'true',
          'x-opencode-directory': REPO_PATH,
        },
      }
    },
  }

  experimental_workspace.register('daytona', adaptor)

  return {}
}

export default DaytonaWorkspacePlugin
