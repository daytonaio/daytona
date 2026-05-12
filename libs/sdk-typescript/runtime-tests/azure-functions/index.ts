// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { app, HttpRequest, HttpResponseInit, InvocationContext } from '@azure/functions'
import { Daytona, Image } from '@daytona/sdk'

export async function sandboxesHandler(_req: HttpRequest, _ctx: InvocationContext): Promise<HttpResponseInit> {
  const image = Image.base('alpine').env({ FOO: 'bar' })
  const daytona = new Daytona({
    apiKey: process.env.DAYTONA_API_KEY,
    apiUrl: process.env.DAYTONA_API_URL,
  })
  const r: any = await daytona.list()
  return {
    jsonBody: {
      imageOk: image.dockerfile.includes('FROM alpine'),
      listOk: Array.isArray(r.items),
    },
  }
}

app.http('sandboxes', { methods: ['GET'], handler: sandboxesHandler })
