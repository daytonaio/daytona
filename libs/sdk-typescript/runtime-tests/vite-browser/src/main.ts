// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Daytona, Image } from '@daytona/sdk'

const result = {
  imageOk: false,
  daytonaConstructorOk: false,
  fsThrowsOk: false,
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

;(window as any).__runtimeTestResult = result
document.body.setAttribute('data-result', JSON.stringify(result))
console.log('RUNTIME_TEST_RESULT:' + JSON.stringify(result))
