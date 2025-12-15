/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Region } from '@daytonaio/api-client'
import { createContext } from 'react'

export interface IRegionsContext {
  loadingRegions: boolean
  sharedRegions: Region[]
  availableRegions: Region[]
  customRegions: Region[]
  refreshAvailableRegions: () => Promise<Region[]>
  getRegionName: (regionId: string) => string | undefined
}

export const RegionsContext = createContext<IRegionsContext | undefined>(undefined)
