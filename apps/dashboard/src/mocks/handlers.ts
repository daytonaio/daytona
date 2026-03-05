/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DaytonaConfiguration } from '@daytonaio/api-client/src'
import { bypass, http, HttpResponse } from 'msw'

const API_URL = import.meta.env.VITE_API_URL

export const handlers = [
  http.get(`${API_URL}/config`, async () => {
    const originalConfig = await fetch(bypass(`${API_URL}/config`)).then((res) => res.json())

    return HttpResponse.json<Partial<DaytonaConfiguration>>({
      ...originalConfig,
    })
  }),
]
