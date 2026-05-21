// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Daytona, Image } from '@daytona/sdk'

export const dynamic = 'force-dynamic'

export async function GET() {
  const image = Image.base('alpine').env({ FOO: 'bar' })
  const daytona = new Daytona()
  const r = await daytona.snapshot.list()
  return Response.json({
    imageOk: image.dockerfile.includes('FROM alpine'),
    listOk: Array.isArray(r.items),
  })
}
