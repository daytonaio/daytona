// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Daytona } from '../../../src/Daytona'
import { Image } from '../../../src/Image'

export async function run() {
  const image = Image.base('alpine').env({ FOO: 'bar' })
  if (!image.dockerfile.includes('FROM alpine')) throw new Error('Image.base failed')

  const daytona = new Daytona()
  const r: any = await daytona.list()
  if (!Array.isArray(r.items)) throw new Error('list() did not return items array')
  return 'PASS'
}
