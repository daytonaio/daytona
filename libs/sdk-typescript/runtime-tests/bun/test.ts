// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { Daytona, Image } from '@daytona/sdk'

const image = Image.base('alpine').env({ FOO: 'bar' })
if (!image.dockerfile.includes('FROM alpine')) throw new Error('Image.base failed')
if (!image.dockerfile.includes('ENV FOO')) throw new Error('Image.env failed')

const daytona = new Daytona()
const iter = daytona.list()
if (typeof (iter as any)[Symbol.asyncIterator] !== 'function') {
  throw new Error('list() did not return an async iterator')
}
const first = await iter.next()
if (typeof first !== 'object' || !('done' in first)) {
  throw new Error('list() iterator did not yield a valid result')
}
console.log('PASS')
