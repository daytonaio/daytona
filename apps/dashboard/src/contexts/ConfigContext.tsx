/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DaytonaConfiguration } from '@daytonaio/api-client'
import { createContext } from 'react'

export const ConfigContext = createContext<DaytonaConfiguration | null>(null)
