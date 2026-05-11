// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

/**
 * SDK contract tests for daytona.session.*
 *
 * Each test starts as `it.skip(...)` until the corresponding implementation lands.
 * Once unskipped, runs against a real Daytona API configured via DAYTONA_API_URL
 * + DAYTONA_API_KEY (same env contract as the existing e2e.test.ts).
 *
 * Smaller scope than the Go suite — only validates that the SDK correctly surfaces
 * the API contract: ergonomics, typed-error mapping, WebSocket abort/cleanup.
 */

jest.setTimeout(120000)

const ENV_OK = !!process.env.DAYTONA_API_URL && !!process.env.DAYTONA_API_KEY

// Marker used by every test until production code lands. Once each implementation todo
// is complete, the corresponding `it.skip` becomes a real `it`.
const todo = (name: string, _fn: () => unknown | Promise<unknown>) => it.skip(name, () => undefined)

describe('Session SDK contract (real API)', () => {
  if (!ENV_OK) {
    it('skipped: DAYTONA_API_URL / DAYTONA_API_KEY not set', () => {
      // intentional no-op so the suite reports as a single skip rather than failing at import
    })
    return
  }

  todo('daytona.session.run() returns { stdout, durationMs, error: null, displays: [] }', async () => {
    // const daytona = new Daytona()
    // const result = await daytona.session.run("print(1)")
    // expect(result.stdout).toBe('1\n')
    // expect(result.durationMs).toBeGreaterThan(0)
    // expect(result.error).toBeNull()
    // expect(result.displays).toEqual([])
  })

  todo('daytona.session.runStream(...) invokes onStdout, onError, onDisplay in arrival order', async () => {
    // const daytona = new Daytona()
    // const events: string[] = []
    // const final = await daytona.session.runStream(
    //   "print('a')\nprint('b')\nimport pandas as pd; pd.DataFrame({'x':[1]})\n1/0",
    //   {
    //     language: 'python',
    //     onStdout: chunk => events.push('stdout:' + chunk),
    //     onError: err => events.push('error:' + err.name),
    //     onDisplay: d => events.push('display:' + d.formats.join(',')),
    //   },
    // )
    // expect(events).toContain('stdout:a\n')
    // expect(events.find(e => e.startsWith('display:'))).toBeDefined()
    // expect(events.find(e => e.startsWith('error:ZeroDivisionError'))).toBeDefined()
    // expect(final.stdout).toBe('a\nb\n')
  })

  todo('createSession + run({context}) (no template/language) succeeds; second call sees state', async () => {
    // const daytona = new Daytona()
    // const ctx = await daytona.session.createSession({ template: 'python-default', language: 'python' })
    // try {
    //   await daytona.session.run('x = 42', { context: ctx })
    //   const r = await daytona.session.run('print(x)', { context: ctx })
    //   expect(r.stdout).toBe('42\n')
    // } finally {
    //   await daytona.session.deleteSession(ctx)
    // }
  })

  todo('410 SessionInvalidatedError is thrown as a typed exception with { sessionId, invalidatedAt }', async () => {
    // const daytona = new Daytona()
    // const ctx = await daytona.session.createSession({ template: 'python-default', language: 'python' })
    // // Force invalidation via test infra (e.g. stopping the underlying sandbox)
    // // Then:
    // await expect(daytona.session.run('print(1)', { context: ctx })).rejects.toMatchObject({
    //   name: 'SessionInvalidatedError',
    //   sessionId: ctx.id,
    //   invalidatedAt: expect.any(String),
    // })
  })

  todo('autoRecreateOnInvalidation: true transparently retries with a fresh context', async () => {
    // const daytona = new Daytona({ apiKey: '...' /* construct flag-bearing instance */ })
    // // Configure: daytona.session = new SessionService(..., { autoRecreateOnInvalidation: true })
    // // Then run with an invalidated context — should silently succeed.
  })

  todo('410 SessionExpiredError is thrown with { sessionId, expiredAt, reason }', async () => {
    // const daytona = new Daytona()
    // const ctx = await daytona.session.createSession({ template: 'python-default', language: 'python' })
    // // Wait for idle TTL (test override required)
    // await expect(daytona.session.run('print(1)', { context: ctx })).rejects.toMatchObject({
    //   name: 'SessionExpiredError',
    //   sessionId: ctx.id,
    //   expiredAt: expect.any(String),
    //   reason: 'idle',
    // })
  })

  todo('AbortSignal cleanly aborts runStream and cleans up the auto-created context', async () => {
    // const daytona = new Daytona()
    // const controller = new AbortController()
    // const promise = daytona.session.runStream('import time; [time.sleep(1) for _ in range(60)]', {
    //   language: 'python',
    //   signal: controller.signal,
    //   onStdout: () => undefined,
    // })
    // setTimeout(() => controller.abort(), 200)
    // await expect(promise).rejects.toMatchObject({ name: 'AbortError' })
    // // List contexts and assert no orphan from this run.
    // const list = await daytona.session.listSessions()
    // expect(list.length).toBeLessThan(50) // smoke: no leak
  })
})
