/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * OpenCode plugin that registers Daytona sandboxes as a workspace adaptor.
 * Each session spawns a remote sandbox running `opencode serve`; tool calls
 * are proxied over the preview URL rather than invoked locally.
 */

import type { PluginInput } from '@opencode-ai/plugin'
import { spawn as nodeSpawn } from 'node:child_process'
import { mkdir, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { Daytona } from '@daytona/sdk'
import type { Sandbox } from '@daytona/sdk'

// Lazy so DAYTONA_API_KEY is read at use-time, not module-load time.
let client: Daytona | undefined

function getDaytona(): Daytona {
  if (client == null) {
    client = new Daytona({
      apiKey: process.env.DAYTONA_API_KEY,
    })
  }
  return client
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

async function withSandbox<T>(name: string, fn: (sandbox: Sandbox) => Promise<T>): Promise<T> {
  const sandbox = await getDaytona().get(sandboxName(name))
  return fn(sandbox)
}

export const DaytonaWorkspacePlugin = async (input: PluginInput) => {
  const { experimental_workspace, worktree, project } = input
  experimental_workspace.register('daytona', {
    name: 'Daytona',
    description: 'Create a remote Daytona sandbox workspace',

    // No-op: opencode's default config is fine as-is.
    configure(config) {
      return config
    },

    // Provision a fresh sandbox: upload the repo, install opencode, start `opencode serve`.
    async create(config) {
      const temp = join(tmpdir(), `opencode-daytona-${Date.now()}`)

      try {
        const sandbox = await getDaytona().create({
          name: sandboxName(config.name),
        })

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
        await spawnAsync(['tar', '-czf', tar, '-C', temp, 'repo'])

        await sandbox.fs.uploadFile(tar, 'repo.tgz')
        await run(
          `rm -rf ${sh(REPO_PATH)} && mkdir -p ${sh(ROOT_PATH)} && tar -xzf "$HOME/repo.tgz" -C "$HOME/workspace"`,
        )

        await run(
          `mkdir -p "$HOME/.opencode/bin" && OPENCODE_INSTALL_DIR="$HOME/.opencode/bin" curl -fsSL https://opencode.ai/install | bash`,
        )

        await sandbox.fs.uploadFile(Buffer.from(`${project.id}\n`), `${REPO_PATH}/.git/opencode`)

        // Create instructions file for Daytona sandbox context
        const sandboxId = sandbox.id
        const instructions = `## Daytona Sandbox Integration
This session is integrated with a Daytona sandbox.
The main project repository is located at: ${REPO_PATH}

### Running Servers
When starting long-running processes like servers, use \`nohup\` to prevent them from being killed when the bash command times out:
\`\`\`bash
nohup <command> > /tmp/server.log 2>&1 &
\`\`\`
For example:
\`\`\`bash
nohup python3 -m http.server 8000 > /tmp/http-server.log 2>&1 &
\`\`\`

### Preview URLs
Before showing a preview URL, ensure the server is running in the sandbox on that port.
To access a running server from a browser, use the Daytona proxy URL format:
\`\`\`
https://<port>-${sandboxId}.daytonaproxy01.net/
\`\`\`
For example, if a server is running on port 8000:
\`\`\`
https://8000-${sandboxId}.daytonaproxy01.net/
\`\`\`
`
        await sandbox.fs.uploadFile(Buffer.from(instructions), `${REPO_PATH}/.opencode/instructions/daytona.md`)

        // Create opencode.json to load the instructions
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
