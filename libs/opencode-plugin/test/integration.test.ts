/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Integration test that runs OpenCode with the Daytona plugin
 * and tests workspace creation via the API.
 */

import { afterAll, beforeAll, describe, expect, test } from 'bun:test'
import { spawn, ChildProcess } from 'node:child_process'
import { mkdir, writeFile, rm, readdir } from 'node:fs/promises'
import { createServer } from 'node:net'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { Daytona } from '@daytona/sdk'

const OPENCODE_BIN = process.env.OPENCODE_BIN || `${process.env.HOME}/.opencode/bin/opencode`
const HAS_DAYTONA_KEY = Boolean(process.env.DAYTONA_API_KEY)

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

// Pick an OS-assigned free port so concurrent opencode servers don't collide.
async function freePort(): Promise<number> {
  return await new Promise((resolve, reject) => {
    const srv = createServer()
    srv.unref()
    srv.on('error', reject)
    srv.listen(0, '127.0.0.1', () => {
      const addr = srv.address()
      if (typeof addr !== 'object' || addr === null) {
        srv.close()
        reject(new Error('failed to obtain port'))
        return
      }
      const port = addr.port
      srv.close(() => resolve(port))
    })
  })
}

async function spawnAsync(cmd: string[], options: { cwd?: string } = {}): Promise<string> {
  return new Promise((resolve, reject) => {
    const proc = spawn(cmd[0], cmd.slice(1), {
      cwd: options.cwd,
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    let stdout = ''
    let stderr = ''
    proc.stdout?.on('data', (data: Buffer) => {
      stdout += data.toString()
    })
    proc.stderr?.on('data', (data: Buffer) => {
      stderr += data.toString()
    })
    proc.on('close', (code: number | null) => {
      if (code === 0) resolve(stdout)
      else reject(new Error(stderr || `Exit code ${code}`))
    })
    proc.on('error', reject)
  })
}

async function waitForServer(port: number, maxWait = 60000): Promise<boolean> {
  const start = Date.now()
  while (Date.now() - start < maxWait) {
    try {
      const res = await fetch(`http://127.0.0.1:${port}/global/health`)
      if (res.ok) return true
    } catch {
      // Server not ready yet
    }
    await sleep(500)
  }
  return false
}

async function createTestProject(baseDir: string): Promise<string> {
  const projectDir = join(baseDir, 'test-project')
  await mkdir(projectDir, { recursive: true })

  // Create test files
  await writeFile(join(projectDir, 'README.md'), '# Integration Test Project\n')
  await writeFile(join(projectDir, 'index.ts'), 'export const hello = () => "Hello!";\n')
  await writeFile(join(projectDir, 'package.json'), JSON.stringify({ name: 'test', version: '1.0.0' }, null, 2))

  // Create src directory
  await mkdir(join(projectDir, 'src'))
  await writeFile(join(projectDir, 'src', 'main.ts'), 'console.log("main");\n')

  // Initialize git
  await spawnAsync(['git', 'init'], { cwd: projectDir })
  await spawnAsync(['git', 'config', 'user.email', 'test@test.com'], { cwd: projectDir })
  await spawnAsync(['git', 'config', 'user.name', 'Test'], { cwd: projectDir })
  await spawnAsync(['git', 'add', '-A'], { cwd: projectDir })
  await spawnAsync(['git', 'commit', '-m', 'init'], { cwd: projectDir })

  // Copy plugin as a single file at .opencode/plugin/daytona.ts
  const pluginDir = join(projectDir, '.opencode', 'plugin')
  await mkdir(pluginDir, { recursive: true })

  // Create a simplified inline version of the plugin for testing
  const pluginContent = `
import { Daytona } from '@daytona/sdk'
import type { PluginInput, WorkspaceAdapter } from '@opencode-ai/plugin'

const REPO_PATH = '/home/daytona/workspace/repo'
const SERVER_PORT = 3096
const HEALTH_URL = \`http://127.0.0.1:\${SERVER_PORT}/global/health\`

function sh(value: string): string {
  return "'" + value.replace(/'/g, "'\\"'\\"'") + "'"
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

let daytonaClient: Daytona | undefined

function getDaytona(): Daytona {
  if (daytonaClient == null) {
    daytonaClient = new Daytona({ apiKey: process.env.DAYTONA_API_KEY })
  }
  return daytonaClient
}

const previewCache = new Map<string, { url: string; token: string }>()

function sandboxName(name: string): string {
  return \`opencode-\${name}\`
}

export const DaytonaPlugin = async (input: PluginInput) => {
  const { experimental_workspace } = input

  if (!process.env.DAYTONA_API_KEY) {
    console.warn('[daytona] DAYTONA_API_KEY is not set')
    return {}
  }

  const adaptor: WorkspaceAdapter = {
    name: 'Daytona',
    description: 'Create a remote Daytona sandbox workspace',

    configure(config) {
      return config
    },

    async create(config) {
      console.log('[daytona] Creating workspace:', config.name)

      const d = getDaytona()
      const sandbox = await d.create({ name: sandboxName(config.name) })

      // Create directory structure
      await sandbox.process.executeCommand(\`mkdir -p \${sh(REPO_PATH)}\`)

      // Install and start opencode
      await sandbox.process.executeCommand(
        \`mkdir -p "$HOME/.opencode/bin" && OPENCODE_INSTALL_DIR="$HOME/.opencode/bin" curl -fsSL https://opencode.ai/install | bash\`
      )

      await sandbox.process.executeCommand(
        \`cd \${sh(REPO_PATH)} && nohup "$HOME/.opencode/bin/opencode" serve --hostname 0.0.0.0 --port \${SERVER_PORT} >/tmp/opencode.log 2>&1 </dev/null &\`
      )

      // Wait for server
      for (let i = 0; i < 60; i++) {
        const result = await sandbox.process.executeCommand(\`curl -fsS \${sh(HEALTH_URL)}\`)
        if (result.exitCode === 0) {
          console.log('[daytona] Server ready')
          return
        }
        await sleep(1000)
      }

      throw new Error('Server did not start')
    },

    async remove(config) {
      const d = getDaytona()
      const sandbox = await d.get(sandboxName(config.name)).catch(() => undefined)
      if (!sandbox) return
      await d.delete(sandbox)
      previewCache.delete(config.name)
    },

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
  console.log('[daytona] Registered daytona adapter')

  return {}
}

export default DaytonaPlugin
`
  await writeFile(join(pluginDir, 'daytona.ts'), pluginContent)

  // Install dependencies needed by the plugin
  await writeFile(
    join(projectDir, 'package.json'),
    JSON.stringify(
      {
        name: 'integration-test',
        version: '1.0.0',
        dependencies: {
          '@daytona/sdk': '*',
          '@opencode-ai/plugin': '*',
        },
      },
      null,
      2,
    ),
  )

  console.log('Installing plugin dependencies...')
  await spawnAsync(['npm', 'install'], { cwd: projectDir })

  // Create opencode.json config
  await writeFile(
    join(projectDir, 'opencode.json'),
    JSON.stringify({
      $schema: 'https://opencode.ai/config.json',
    }),
  )

  return projectDir
}

describe.skipIf(!HAS_DAYTONA_KEY)('integration', () => {
  let testDir: string
  let projectDir: string
  let serverProc: ChildProcess | null = null
  let serverPort: number
  let createdSandboxId: string | null = null
  let daytona: Daytona

  beforeAll(async () => {
    daytona = new Daytona({ apiKey: process.env.DAYTONA_API_KEY })
    serverPort = await freePort()
    testDir = join(tmpdir(), `integration-test-${Date.now()}`)

    console.log('\n=== Step 1: Create test project with plugin ===')
    await mkdir(testDir, { recursive: true })
    projectDir = await createTestProject(testDir)
    console.log(`Project created at: ${projectDir}`)

    const pluginCheck = await readdir(join(projectDir, '.opencode', 'plugin'))
    console.log(`Plugin files: ${pluginCheck.join(', ')}`)

    console.log('\n=== Step 2: Start OpenCode server ===')
    console.log(`Running: ${OPENCODE_BIN} serve --port ${serverPort}`)

    serverProc = spawn(OPENCODE_BIN, ['serve', '--port', String(serverPort)], {
      cwd: projectDir,
      env: {
        ...process.env,
        OPENCODE_EXPERIMENTAL_WORKSPACES: 'true',
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    })

    serverProc.stdout?.on('data', (d: Buffer) => {
      process.stdout.write(d)
    })
    serverProc.stderr?.on('data', (d: Buffer) => {
      process.stderr.write(d)
    })

    console.log('Waiting for server...')
    const serverReady = await waitForServer(serverPort)
    if (!serverReady) {
      throw new Error('Server did not start')
    }
    console.log('Server is ready!')
  })

  afterAll(async () => {
    console.log('\n=== Cleanup ===')

    if (serverProc) {
      console.log('Stopping server...')
      serverProc.kill('SIGTERM')
    }

    if (createdSandboxId) {
      console.log('Deleting sandbox...')
      try {
        const sandbox = await daytona.get(createdSandboxId)
        await daytona.delete(sandbox)
      } catch {
        // Ignore cleanup errors
      }
    }

    console.log('Removing test directory...')
    await rm(testDir, { recursive: true, force: true }).catch(() => undefined)
  })

  test('registers daytona workspace adapter', async () => {
    console.log('\n=== Step 3: Check workspace adapters ===')
    const adaptersRes = await fetch(`http://127.0.0.1:${serverPort}/experimental/workspace/adapter`)
    const adapters = await adaptersRes.json()
    console.log('Adapters:', JSON.stringify(adapters, null, 2))

    expect(Array.isArray(adapters)).toBe(true)
    expect(adapters.some((a: { type: string }) => a.type === 'daytona')).toBe(true)
  })

  test('creates and deletes workspace via API', async () => {
    console.log('\n=== Step 4: Create workspace via API ===')
    const createRes = await fetch(`http://127.0.0.1:${serverPort}/experimental/workspace`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        type: 'daytona',
        name: `integration-test-${Date.now()}`,
        branch: 'master',
      }),
    })

    console.log(`Create response status: ${createRes.status}`)
    const createBody = await createRes.text()
    console.log(`Create response body: ${createBody}`)

    expect(createRes.ok).toBe(true)

    const workspace = JSON.parse(createBody)
    console.log('Workspace created:', workspace)

    // Step 5: Verify sandbox
    console.log('\n=== Step 5: Verify sandbox contents ===')
    const sandboxName = `opencode-${workspace.name}`
    await sleep(5000)

    const sandbox = await daytona.get(sandboxName)
    createdSandboxId = sandbox.id

    const files = await sandbox.process.executeCommand('ls -la /home/daytona/workspace/repo')
    console.log('Sandbox files:', files.result)
    expect(files.exitCode).toBe(0)

    // Step 6: Clean up workspace
    console.log('\n=== Step 6: Clean up workspace ===')
    const deleteRes = await fetch(`http://127.0.0.1:${serverPort}/experimental/workspace/${workspace.id}`, {
      method: 'DELETE',
    })
    console.log(`Delete response: ${deleteRes.status}`)
    expect(deleteRes.ok).toBe(true)

    // Mark as cleaned up so afterAll doesn't try again
    createdSandboxId = null
  })
})
