/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ReactNode, useCallback, useEffect, useMemo, useState } from 'react'
import { Region, RegionType } from '@daytonaio/api-client'
import { IRegionsContext, RegionsContext } from '@/contexts/RegionsContext'
import { useApi } from '@/hooks/useApi'
import { handleApiError } from '@/lib/error-handling'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'

type Props = {
  children: ReactNode
}

export function RegionsProvider(props: Props) {
  const { regionsApi } = useApi()

  const { selectedOrganization } = useSelectedOrganization()

  const [regions, setRegions] = useState<Region[]>([])
  const [loadingRegions, setLoadingRegions] = useState(true)

  const getRegions = useCallback(async () => {
    if (!selectedOrganization) {
      setRegions([])
      setLoadingRegions(false)
      return []
    }
    try {
      const regions = (await regionsApi.listRegions(selectedOrganization.id)).data
      setRegions(regions)
      return regions
    } catch (error) {
      handleApiError(error, 'Failed to fetch regions')
      setRegions([])
      throw error
    } finally {
      setLoadingRegions(false)
    }
  }, [regionsApi, selectedOrganization])

  useEffect(() => {
    getRegions()
  }, [getRegions])

  const getRegionName = useCallback(
    (regionId: string): string | undefined => {
      const region = regions.find((region) => region.id === regionId)
      return region?.name
    },
    [regions],
  )

  const sharedRegions = useMemo(() => {
    return regions.filter((region) => region.regionType === RegionType.SHARED)
  }, [regions])

  const customRegions = useMemo(() => {
    return regions.filter((region) => region.regionType === RegionType.CUSTOM)
  }, [regions])

  const contextValue: IRegionsContext = useMemo(() => {
    return {
      regions,
      loadingRegions,
      sharedRegions,
      customRegions,
      refreshRegions: getRegions,
      getRegionName,
    }
  }, [regions, loadingRegions, sharedRegions, customRegions, getRegions, getRegionName])

  return <RegionsContext.Provider value={contextValue}>{props.children}</RegionsContext.Provider>
}
