// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Daytona, Image } from '@daytona/sdk'

export const handler = async () => {
  const image = Image.base('alpine').env({ FOO: 'bar' })
  const daytona = new Daytona()
  const r: any = await daytona.list()
  return {
    statusCode: 200,
    body: JSON.stringify({
      imageOk: image.dockerfile.includes('FROM alpine'),
      listOk: Array.isArray(r.items),
    }),
  }
}
