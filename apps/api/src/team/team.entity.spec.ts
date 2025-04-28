/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Team } from './team.entity'

describe('Team', () => {
  it('should be defined', () => {
    expect(new Team()).toBeDefined()
  })
})
