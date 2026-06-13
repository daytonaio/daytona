/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BASE_PROPAGATION_FACTOR } from '../constants/snapshot.constants'
import { isColdSnapshot } from './snapshot.util'

describe('isColdSnapshot', () => {
  it('returns true when propagationFactor is 0', () => {
    expect(isColdSnapshot({ propagationFactor: 0 })).toBe(true)
  })

  it('returns false for the default warm propagation factor', () => {
    expect(isColdSnapshot({ propagationFactor: BASE_PROPAGATION_FACTOR })).toBe(false)
  })

  it('returns false for any positive propagation factor', () => {
    expect(isColdSnapshot({ propagationFactor: 1 })).toBe(false)
    expect(isColdSnapshot({ propagationFactor: 0.1 })).toBe(false)
  })
})
