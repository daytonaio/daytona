// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

const { Daytona, Image } = require('@daytona/sdk')

const image = Image.base('alpine').env({ FOO: 'bar' })
if (!image.dockerfile.includes('FROM alpine')) throw new Error('Image.base failed')
if (!image.dockerfile.includes('ENV FOO')) throw new Error('Image.env failed')

const daytona = new Daytona()
daytona
  .list()
  .then((r) => {
    if (!Array.isArray(r.items)) throw new Error('list() did not return items array')
    console.log('PASS')
  })
  .catch((e) => {
    console.error('FAIL:', e.message)
    process.exit(1)
  })