// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Daytona } from '../../../../src/Daytona'
import { Image } from '../../../../src/Image'

export default defineEventHandler(async () => {
  const image = Image.base('alpine').env({ FOO: 'bar' })
  const daytona = new Daytona()
  const r: any = await daytona.list()
  return {
    imageOk: image.dockerfile.includes('FROM alpine'),
    listOk: Array.isArray(r.items),
  }
})
