// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { json } from '@remix-run/node'
import { Daytona, Image } from '@daytona/sdk'

export async function loader() {
  const image = Image.base('alpine').env({ FOO: 'bar' })
  const daytona = new Daytona()
  const r: any = await daytona.list()
  return json({
    imageOk: image.dockerfile.includes('FROM alpine'),
    listOk: Array.isArray(r.items),
  })
}
