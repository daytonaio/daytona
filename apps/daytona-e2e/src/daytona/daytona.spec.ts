/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios from 'axios'

describe('GET /api', () => {
  it('should return a message', async () => {
    const res = await axios.get(`/api`)

    expect(res.status).toBe(200)
    expect(res.data).toEqual({ message: 'Hello API' })
  })
})
