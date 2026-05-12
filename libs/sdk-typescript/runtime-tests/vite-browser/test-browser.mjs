import { chromium } from 'playwright'
import { createServer } from 'http'
import { readFileSync, existsSync } from 'fs'
import { extname, join, resolve } from 'path'

const DIST = resolve('dist')

const mime = { '.js': 'application/javascript', '.html': 'text/html', '.css': 'text/css', '.json': 'application/json' }

const server = createServer((req, res) => {
  let p = req.url.split('?')[0]
  if (p === '/') p = '/index.html'
  const fp = resolve(join(DIST, p))
  if (!fp.startsWith(DIST + '/') && fp !== DIST) {
    res.writeHead(403)
    res.end()
    return
  }
  if (existsSync(fp)) {
    res.writeHead(200, { 'content-type': mime[extname(fp)] || 'text/plain' })
    res.end(readFileSync(fp))
  } else {
    res.writeHead(404)
    res.end()
  }
})

await new Promise((r) => server.listen(0, r))
const PORT = server.address().port

const cleanup = () => {
  try { server.close() } catch { /* noop */ }
}
process.on('exit', cleanup)

const browser = await chromium.launch({ headless: true })
try {
  const page = await browser.newPage()
  const errors = []
  page.on('pageerror', (e) => errors.push('pageerror: ' + e.message))
  page.on('console', (msg) => { if (msg.type() === 'error') errors.push('console.error: ' + msg.text()) })

  const allConsole = []
  page.on('console', (msg) => allConsole.push(`[${msg.type()}] ${msg.text()}`))
  page.on('pageerror', (e) => allConsole.push(`[pageerror] ${e.message}`))

  await page.goto(`http://localhost:${PORT}/`, { waitUntil: 'networkidle' })
  await page.waitForFunction(() => document.body.hasAttribute('data-result'), { timeout: 5000 }).catch(() => {})

  const body = await page.locator('body').first()
  const attr = await body.getAttribute('data-result')
  if (!attr) {
    console.error('FAIL: no data-result attribute')
    console.error('HTML:', await page.content())
    console.error('Browser console:', allConsole)
    process.exit(1)
  }
  const result = JSON.parse(attr)

  console.log('Result:', JSON.stringify(result))
  if (errors.length) console.log('Browser errors:', errors)

  if (!result.imageOk) { console.error('FAIL: Image API broken'); process.exit(1) }
  if (!result.daytonaConstructorOk) { console.error('FAIL: Daytona constructor broken'); process.exit(1) }
  if (!result.fsThrowsOk) { console.error('FAIL: fs methods did not throw clear error'); process.exit(1) }

  console.log('PASS')
} finally {
  await browser.close()
  cleanup()
}
