// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { chromium } from 'playwright'
import { createServer } from 'http'
import { readFileSync, existsSync } from 'fs'
import { extname, join, resolve } from 'path'
import { Daytona } from '@daytona/sdk'

const DIST = resolve('dist')
const mime = { '.js': 'application/javascript', '.html': 'text/html', '.css': 'text/css', '.json': 'application/json' }

const API_KEY = process.env.DAYTONA_API_KEY
const API_URL = process.env.DAYTONA_API_URL
if (!API_KEY || !API_URL) {
  console.error('FAIL: DAYTONA_API_KEY and DAYTONA_API_URL must be set')
  process.exit(1)
}

const FILE_CONTENT = 'hello buffer'
const FILE_PATH = 'test.txt'

// ---------------------------------------------------------------------------
// Create sandbox + upload test file (Node.js SDK, runs before browser launch)
// ---------------------------------------------------------------------------

console.log('Creating sandbox...')
const daytona = new Daytona({ apiKey: API_KEY, apiUrl: API_URL })
const sandbox = await daytona.create({ timeout: 120, labels: { purpose: 'runtime-test-vite-browser' } })
console.log('Sandbox created:', sandbox.id)

let exitCode = 0

try {
  await sandbox.fs.uploadFile(Buffer.from(FILE_CONTENT), FILE_PATH)
  console.log('Uploaded test file:', FILE_PATH)

  exitCode = await runBrowserTest(sandbox.id)
} finally {
  console.log('Deleting sandbox:', sandbox.id)
  await daytona.delete(sandbox).catch((e) => console.warn('Sandbox delete warning:', e.message))
}

process.exit(exitCode)

// ---------------------------------------------------------------------------
// Browser test — isolated in a function so we can use early returns cleanly
// ---------------------------------------------------------------------------

async function runBrowserTest(sandboxId) {
  // Injects Daytona config into index.html at serve time so the browser bundle
  // can reach the real API and sandbox toolbox.
  // Both http://localhost:3001/api and http://proxy.localhost:4000/toolbox
  // support CORS with reflected origins, so no proxy is needed.
  const injectedScript = `<script>
    window.__DAYTONA_API_KEY__ = ${JSON.stringify(API_KEY)};
    window.__DAYTONA_API_URL__ = ${JSON.stringify(API_URL)};
    window.__TEST_SANDBOX_ID__ = ${JSON.stringify(sandboxId)};
    window.__TEST_FILE_CONTENT__ = ${JSON.stringify(FILE_CONTENT)};
  </script>`

  const server = createServer((req, res) => {
    let p = req.url.split('?')[0]
    if (p === '/') p = '/index.html'
    const fp = resolve(join(DIST, p))
    if (!fp.startsWith(DIST + '/') && fp !== DIST) {
      res.writeHead(403); res.end(); return
    }
    if (existsSync(fp)) {
      if (extname(fp) === '.html') {
        const html = readFileSync(fp, 'utf-8').replace('<head>', `<head>${injectedScript}`)
        res.writeHead(200, { 'content-type': 'text/html' })
        res.end(html)
      } else {
        res.writeHead(200, { 'content-type': mime[extname(fp)] || 'text/plain' })
        res.end(readFileSync(fp))
      }
    } else {
      res.writeHead(404); res.end()
    }
  })

  await new Promise((r) => server.listen(0, r))
  const PORT = server.address().port
  const closeServer = () => { try { server.close() } catch { /* noop */ } }

  const browser = await chromium.launch({ headless: true })
  try {
    const page = await browser.newPage()
    const errors = []
    const allConsole = []
    page.on('pageerror', (e) => { errors.push('pageerror: ' + e.message); allConsole.push(`[pageerror] ${e.message}`) })
    page.on('console', (msg) => {
      allConsole.push(`[${msg.type()}] ${msg.text()}`)
      if (msg.type() === 'error') errors.push('console.error: ' + msg.text())
    })

    await page.goto(`http://localhost:${PORT}/`, { waitUntil: 'networkidle' })
    await page.waitForFunction(() => document.body.hasAttribute('data-result'), { timeout: 15000 }).catch(() => {})

    const attr = await page.locator('body').first().getAttribute('data-result')
    if (!attr) {
      console.error('FAIL: no data-result attribute set')
      console.error('HTML:', await page.content())
      console.error('Browser console:', allConsole)
      return 1
    }

    const result = JSON.parse(attr)
    console.log('Result:', JSON.stringify(result))
    if (errors.length) console.log('Browser errors:', errors)

    const checks = [
      ['imageOk', 'Image API broken'],
      ['daytonaConstructorOk', 'Daytona constructor broken'],
      ['fsThrowsOk', 'fs methods did not throw clear error'],
      ['bufferOk', 'globalThis.Buffer polyfill not available or broken'],
      ['listOk', 'daytona.list() failed in browser'],
      ['downloadFileOk', 'FileSystem.downloadFile Buffer handling broken in browser (Binary.ts regression)'],
    ]

    let code = 0
    for (const [key, msg] of checks) {
      if (!result[key]) { console.error(`FAIL: ${msg}`); code = 1 }
    }
    if (code === 0) console.log('PASS')
    return code
  } finally {
    await browser.close()
    closeServer()
  }
}
