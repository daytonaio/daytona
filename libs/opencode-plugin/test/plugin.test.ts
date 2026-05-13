/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterAll, beforeAll, describe, expect, test } from 'bun:test'
import { createServer } from 'node:net'
import { resolve } from 'node:path'

import { Daytona } from '@daytona/sdk'
import { createOpencode } from '@opencode-ai/sdk/v2'

const PLUGIN_PATH = resolve(import.meta.dir, '../.opencode/plugin/index.ts')
const PLUGIN_SPEC = `file://${PLUGIN_PATH}`

const HAS_DAYTONA_KEY = Boolean(process.env.DAYTONA_API_KEY)

// Pick an OS-assigned free port so concurrent opencode servers (e.g. the
// developer's own session) don't collide with the test on 4096.
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

describe('opencode-plugin', () => {
  test('registers the daytona workspace adapter', async () => {
    const { client, server } = await createOpencode({
      port: await freePort(),
      timeout: 30_000,
      config: { plugin: [PLUGIN_SPEC] },
    })
    try {
      const { data, error } = await client.experimental.workspace.adapter.list()
      expect(error).toBeUndefined()
      expect(data?.some((a) => a.type === 'daytona')).toBe(true)
    } finally {
      server.close()
    }
  })

  describe.skipIf(!HAS_DAYTONA_KEY)('partial-create cleanup (requires DAYTONA_API_KEY)', () => {
    let daytona: Daytona
    const leakedIds: string[] = []

    beforeAll(() => {
      daytona = new Daytona({ apiKey: process.env.DAYTONA_API_KEY })
    })

    // Belt-and-suspenders: any sandbox the test detected as leaked gets
    // deleted here, so a red test does not also burn money.
    afterAll(async () => {
      for (const id of leakedIds) {
        const sandbox = await daytona.get(id).catch(() => undefined)
        if (sandbox) await daytona.delete(sandbox).catch(() => undefined)
      }
    })

    test('removes sandbox when create() fails partway', async () => {
      // Workspace endpoints are gated behind this opencode feature flag.
      process.env.OPENCODE_EXPERIMENTAL_WORKSPACES = 'true'

      const { client, server } = await createOpencode({
        port: await freePort(),
        timeout: 30_000,
        config: { plugin: [PLUGIN_SPEC] },
      })

      // Snapshot the existing sandbox set so any sandbox that appears after
      // the failed create() is identifiable as a leak. opencode assigns a
      // random petname for `config.name`, so we cannot predict the sandbox
      // name up front.
      const before = new Set((await daytona.list()).items.map((s) => s.id))

      try {
        // The bad branch makes the host-side `git clone --branch ...` fail;
        // by then the plugin has already created the Daytona sandbox.
        const { error } = await client.experimental.workspace.create({
          id: `wrk-e2e-cleanup-${Date.now()}`,
          type: 'daytona',
          branch: `does-not-exist-${Date.now()}`,
          extra: null,
        })

        if (!error) throw new Error('expected workspace.create to fail')

        // Daytona may not surface a freshly-created sandbox in list() right
        // away; poll for up to 30s before declaring no leak.
        const start = Date.now()
        let leaks: Array<{ id: string; name: string }> = []
        while (Date.now() - start < 30_000) {
          const after = (await daytona.list()).items
          leaks = after
            .filter((s) => !before.has(s.id))
            // Daytona renames deleted sandboxes to DESTROYED_<name>_<ts> the
            // moment teardown begins (state may still be "destroying"); those
            // are not leaks — the plugin already cleaned them up.
            .map((s) => ({ id: s.id, name: (s as unknown as { name: string }).name }))
            .filter((s) => !s.name.startsWith('DESTROYED_'))
          if (leaks.length > 0) break
          await new Promise((r) => setTimeout(r, 1000))
        }
        for (const l of leaks) leakedIds.push(l.id)
        console.log(`[partial-create] leaks detected: ${leaks.length} after ${Date.now() - start}ms`, leaks)

        // Sandbox must be gone — leaving it would silently burn money.
        expect(leaks).toEqual([])
      } finally {
        server.close()
      }
    }, 300_000)
  })
})
