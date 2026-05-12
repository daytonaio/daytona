// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Daytona, Image } from '@daytona/sdk'

export default {
  async fetch(_req: Request, env: any) {
    const image = Image.base('alpine').env({ FOO: 'bar' })
    const daytona = new Daytona({
      apiKey: env.DAYTONA_API_KEY,
      apiUrl: env.DAYTONA_API_URL,
    })
    const r: any = await daytona.list()
    return Response.json({
      imageOk: image.dockerfile.includes('FROM alpine'),
      listOk: Array.isArray(r.items),
    })
  },
}
