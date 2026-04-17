/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * OpenCode Plugin: Daytona Workspace Adaptor
 *
 * This plugin integrates Daytona sandboxes with OpenCode using the experimental
 * workspace API. It provisions remote Daytona sandboxes as workspaces, runs the
 * OpenCode server inside the sandbox, and proxies all tool calls through the
 * remote target.
 *
 * Requires:
 * - npm install @daytona/sdk
 * - Environment: DAYTONA_API_KEY
 * - OpenCode flag: OPENCODE_EXPERIMENTAL_WORKSPACES=true
 */

import type { Plugin } from '@opencode-ai/plugin'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { mkdir, rm } from 'node:fs/promises'
import { spawn as nodeSpawn } from 'node:child_process'

// Import SDK types and class
import { Daytona } from '@daytona/sdk'
import type { Sandbox } from '@daytona/sdk'

/** Lazy-initialized Daytona client */
let client: Daytona | undefined

function getDaytona(): Daytona {
  if (client == null) {
    client = new Daytona({
      apiKey: process.env.DAYTONA_API_KEY,
    })
  }
  return client
}

/** Cache for preview links to avoid repeated API calls */
const previewCache = new Map<string, { url: string; token: string }>()

/** Sandbox paths */
const REPO_PATH = '/home/daytona/workspace/repo'
const ROOT_PATH = '/home/daytona/workspace'
const LOCAL_BIN = '/home/daytona/opencode'
const INSTALL_BIN = '/home/daytona/.opencode/bin/opencode'
const HEALTH_URL = 'http://127.0.0.1:3096/global/health'
const SERVER_PORT = 3096

/**
 * Shell-escape a string for safe use in shell commands
 */
function sh(value: string): string {
  return `'${value.replace(/'/g, `'\"'\"'`)}'`
}

/**
 * Sleep for a given number of milliseconds
 */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * Spawn a process and wait for it to exit, throwing on non-zero exit
 */
async function spawnAsync(cmd: string[], options: { cwd?: string } = {}): Promise<void> {
  return new Promise((resolve, reject) => {
    const proc = nodeSpawn(cmd[0], cmd.slice(1), {
      cwd: options.cwd,
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

/**
 * Helper to work with a sandbox by name
 */
async function withSandbox<T>(name: string, fn: (sandbox: Sandbox) => Promise<T>): Promise<T> {
  const sandbox = await getDaytona().get(name)
  return fn(sandbox)
}

/**
 * Convert env object to Record<string, string> for SDK compatibility
 */
function toEnvVars(env: Record<string, unknown>): Record<string, string> {
  const result: Record<string, string> = {}
  for (const [key, value] of Object.entries(env)) {
    if (typeof value === 'string') {
      result[key] = value
    } else if (value != null) {
      result[key] = String(value)
    }
  }
  return result
}

/**
 * Daytona Workspace Plugin
 *
 * Registers a "daytona" workspace type that provisions Daytona sandboxes
 * as remote OpenCode workspaces.
 */
export const DaytonaWorkspacePlugin: Plugin = async ({ experimental_workspace, worktree, project }) => {
  experimental_workspace.register('daytona', {
    name: 'Daytona',
    description: 'Create a remote Daytona sandbox workspace',

    /**
     * Configure workspace metadata (no customization needed)
     */
    configure(config) {
      return config
    },

    /**
     * Create a new Daytona sandbox workspace
     */
    async create(config, env) {
      const temp = join(tmpdir(), `opencode-daytona-${Date.now()}`)

      try {
        // 1. Create the Daytona sandbox
        const sandbox = await getDaytona().create({
          name: config.name,
          envVars: toEnvVars(env as Record<string, unknown>),
        })

        // Helper to run commands in the sandbox
        const run = async (command: string): Promise<void> => {
          const result = await sandbox.process.executeCommand(command)
          if (result.result) {
            process.stdout.write(result.result)
          }
          if (result.exitCode !== 0) {
            throw new Error(result.result || `Sandbox command failed: ${command}`)
          }
        }

        // 2. Clone and package the local repository
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
        await spawnAsync(['tar', '-czf', tar, '-C', temp, 'repo'])

        // 3. Upload repository to sandbox
        await sandbox.fs.uploadFile(tar, 'repo.tgz')

        // 4. Extract repository in sandbox
        await run(`rm -rf ${sh(REPO_PATH)} && mkdir -p ${sh(ROOT_PATH)} && tar -xzf "$HOME/repo.tgz" -C "$HOME/workspace"`)

        // 5. Install OpenCode server
        await run(
          `mkdir -p "$HOME/.opencode/bin" && OPENCODE_INSTALL_DIR="$HOME/.opencode/bin" curl -fsSL https://opencode.ai/install | bash`,
        )

        // 6. Write project ID to .git/opencode for session association
        await run(`printf "%s\\n" ${sh(project.id)} > ${sh(`${REPO_PATH}/.git/opencode`)}`)

        // 7. Start OpenCode server
        await run(
          `cd ${sh(REPO_PATH)} && exe=${sh(LOCAL_BIN)} && if [ ! -x "$exe" ]; then exe=${sh(INSTALL_BIN)}; fi && nohup env "$exe" serve --hostname 0.0.0.0 --port ${SERVER_PORT} >/tmp/opencode.log 2>&1 </dev/null &`,
        )

        // 8. Wait for server to become healthy
        for (let i = 0; i < 60; i++) {
          const result = await sandbox.process.executeCommand(`curl -fsS ${sh(HEALTH_URL)}`)
          if (result.exitCode === 0) {
            if (result.result) {
              process.stdout.write(result.result)
            }
            return
          }
          await sleep(1000)
        }

        // Server didn't start - get logs for debugging
        const log = await sandbox.process.executeCommand('test -f /tmp/opencode.log && cat /tmp/opencode.log || true')
        throw new Error(log.result || 'Daytona workspace server did not become ready in time')
      } finally {
        // Clean up temp directory
        await rm(temp, { recursive: true, force: true }).catch(() => {})
      }
    },

    /**
     * Remove a Daytona sandbox workspace
     */
    async remove(config) {
      const d = getDaytona()
      const sandbox = await d.get(config.name).catch(() => undefined)
      if (!sandbox) return
      await d.delete(sandbox)
      previewCache.delete(config.name)
    },

    /**
     * Get the remote target for proxying tool calls
     */
    async target(config) {
      let link = previewCache.get(config.name)
      if (!link) {
        link = await withSandbox(config.name, (sandbox) => sandbox.getPreviewLink(SERVER_PORT))
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
  })

  return {}
}

export default DaytonaWorkspacePlugin
