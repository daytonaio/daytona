/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Region, RegionType } from '@daytona/api-client'

export const EMPTY_REGIONS: Region[] = []

export const filterCustomRegions = (regions: Region[]) => {
  return regions.filter((region) => region.regionType === RegionType.CUSTOM)
}

export const createRegionNameGetter = (...regionLists: Region[][]) => {
  const regionNameById = new Map<string, string>()

  for (const regions of regionLists) {
    for (const region of regions) {
      if (!regionNameById.has(region.id)) {
        regionNameById.set(region.id, region.name)
      }
    }
  }

  return (regionId: string) => regionNameById.get(regionId)
}
