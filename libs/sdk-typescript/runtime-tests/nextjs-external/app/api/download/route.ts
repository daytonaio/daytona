// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Daytona } from '@daytona/sdk'

export const dynamic = 'force-dynamic'

// Exercises the exact bug path from issue #4771:
//   Next.js App Router (Node runtime) + serverExternalPackages → SDK loaded as
//   ESM → sandbox.fs.downloadFile() → processDownloadFilesResponseWithBusboy →
//   bare `require()` in compiled SDK. Before the fix this throws
//   `ReferenceError: require is not defined`.
export async function GET(request: Request) {
  const url = new URL(request.url)
  const sandboxId = url.searchParams.get('sandboxId')
  const expectedContent = url.searchParams.get('expected')
  if (!sandboxId || !expectedContent) {
    return Response.json({ error: 'sandboxId and expected params required' }, { status: 400 })
  }

  try {
    const daytona = new Daytona()
    const sandbox = await daytona.get(sandboxId)
    const buf = await sandbox.fs.downloadFile('test.txt')
    return Response.json({
      downloadOk: buf.toString('utf-8') === expectedContent,
    })
  } catch (e: unknown) {
    return Response.json({
      downloadOk: false,
      error: e instanceof Error ? e.message : String(e),
      stack: e instanceof Error ? e.stack : undefined,
    })
  }
}
