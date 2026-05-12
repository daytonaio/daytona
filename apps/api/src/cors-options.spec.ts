/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { corsOptions } from './cors-options'

describe('corsOptions', () => {
  it('does not enable credentials with reflected origins', () => {
    expect(corsOptions.credentials).toBe(false)
  })
})
