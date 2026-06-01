// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

// Verifies dynamicRequire still works after esbuild re-bundles the ESM build
// to CJS. pipInstallFromRequirements is a synchronous, no-network path that
// exercises dynamicRequire('fs').

import { Daytona, Image } from '@daytona/sdk'
import * as fs from 'node:fs'
import * as os from 'node:os'
import * as path from 'node:path'

async function main() {
  const image = Image.base('alpine').env({ FOO: 'bar' })
  if (!image.dockerfile.includes('FROM alpine')) throw new Error('Image.base failed')
  if (!image.dockerfile.includes('ENV FOO')) throw new Error('Image.env failed')

  const daytona = new Daytona()
  const iter = daytona.list()
  if (typeof iter[Symbol.asyncIterator] !== 'function') {
    throw new Error('list() did not return an async iterator')
  }
  const first = await iter.next()
  if (typeof first !== 'object' || !('done' in first)) {
    throw new Error('list() iterator did not yield a valid result')
  }

  const reqPath = path.join(fs.mkdtempSync(path.join(os.tmpdir(), 'daytona-esbuild-cjs-')), 'requirements.txt')
  fs.writeFileSync(reqPath, 'requests==2.31.0\n')
  const built = Image.debianSlim('3.12').pipInstallFromRequirements(reqPath)
  if (!built.dockerfile.includes('pip install -r /.requirements.txt')) {
    throw new Error('pipInstallFromRequirements did not produce the expected dockerfile')
  }

  console.log('PASS')
}

main().catch((e) => {
  console.error('FAIL:', e && e.message ? e.message : e)
  process.exit(1)
})
