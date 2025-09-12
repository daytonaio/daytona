/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DashboardConfig } from '@/types/DashboardConfig'
import { createContext } from 'react'

export const ConfigContext = createContext<DashboardConfig | null>(null)
