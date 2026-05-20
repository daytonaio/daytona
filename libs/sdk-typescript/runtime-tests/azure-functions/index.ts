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
  const iter = daytona.list()
  const listOk = typeof (iter as any)[Symbol.asyncIterator] === 'function' && typeof (await iter.next()) === 'object'
  return {
    jsonBody: {
      imageOk: image.dockerfile.includes('FROM alpine'),
      listOk,
    },
  }
}

app.http('sandboxes', { methods: ['GET'], handler: sandboxesHandler })
