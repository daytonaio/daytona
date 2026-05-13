// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Buffer } from 'buffer'
import { Daytona, Image } from '@daytona/sdk'

const result: Record<string, unknown> = {
  imageOk: false,
  daytonaConstructorOk: false,
  fsThrowsOk: false,
  bufferOk: false,
  listOk: false,
  downloadFileOk: false,
}

try {
  const image = Image.base('alpine').env({ FOO: 'bar' })
  result.imageOk = image.dockerfile.includes('FROM alpine') && image.dockerfile.includes('ENV FOO')
} catch {
  result.imageOk = false
}

try {
  new Daytona({ apiKey: 'browser-test', apiUrl: 'http://invalid.example' })
  result.daytonaConstructorOk = true
} catch {
  result.daytonaConstructorOk = false
}

try {
  Image.fromDockerfile('/nonexistent')
  result.fsThrowsOk = false
} catch (e: unknown) {
  const msg = e instanceof Error ? e.message : String(e)
  result.fsThrowsOk = /not available|require.*unavailable|fs/.test(msg)
}

// Verify the Buffer polyfill (vite-plugin-node-polyfills with globals.Buffer:true)
// works correctly — same import pattern used by the dashboard's FileTreePane.tsx.
try {
  const buf = Buffer.from('hello buffer')
  result.bufferOk = buf instanceof Uint8Array && buf.toString('utf-8') === 'hello buffer'
} catch (e: unknown) {
  result.bufferError = e instanceof Error ? e.message : String(e)
  result.bufferOk = false
}

// Real API tests — require DAYTONA_API_KEY / DAYTONA_API_URL injected by the
// Node.js orchestrator and a pre-created sandbox with a test file uploaded.
const apiKey = (window as any).__DAYTONA_API_KEY__ as string | undefined
const apiUrl = (window as any).__DAYTONA_API_URL__ as string | undefined
const sandboxId = (window as any).__TEST_SANDBOX_ID__ as string | undefined
const fileContent = (window as any).__TEST_FILE_CONTENT__ as string | undefined

if (apiKey && apiUrl) {
  const daytona = new Daytona({ apiKey, apiUrl })

  try {
    const r = await daytona.list()
    result.listOk = Array.isArray(r.items)
  } catch (e: unknown) {
    result.listError = e instanceof Error ? e.message : String(e)
    result.listOk = false
  }

  // downloadFile exercises the exact regression path fixed in Binary.ts:
  // processDownloadFilesResponseWithBuffered → toBuffer → getBufferCtor.
  if (sandboxId && fileContent) {
    try {
      const sandbox = await daytona.get(sandboxId)
      const buf = await sandbox.fs.downloadFile('test.txt')
      result.downloadFileOk = buf.toString('utf-8') === fileContent
    } catch (e: unknown) {
      result.downloadFileError = e instanceof Error ? e.message : String(e)
      result.downloadFileOk = false
    }
  }
}

;(window as any).__runtimeTestResult = result
document.body.setAttribute('data-result', JSON.stringify(result))
console.log('RUNTIME_TEST_RESULT:' + JSON.stringify(result))
