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
  const { regionsApi, organizationsApi } = useApi()

  const { selectedOrganization } = useSelectedOrganization()

  const [sharedRegions, setSharedRegions] = useState<Region[]>([])
  const [availableRegions, setAvailableRegions] = useState<Region[]>([])
  const [loadingRegions, setLoadingRegions] = useState(true)

  const getSharedRegions = useCallback(async () => {
    try {
      const regions = (await regionsApi.listSharedRegions()).data
      setSharedRegions(regions)
    } catch (error) {
      handleApiError(error, 'Failed to fetch shared regions')
      setSharedRegions([])
      throw error
    }
  }, [regionsApi])

  const getAvailableRegions = useCallback(async () => {
    if (!selectedOrganization) {
      setAvailableRegions([])
      return []
    }
    try {
      const regions = (await organizationsApi.listAvailableRegions(selectedOrganization.id)).data
      setAvailableRegions(regions)
      return regions
    } catch (error) {
      handleApiError(error, 'Failed to fetch available regions')
      setAvailableRegions([])
      throw error
    }
  }, [organizationsApi, selectedOrganization])

  useEffect(() => {
    Promise.allSettled([getSharedRegions(), getAvailableRegions()]).finally(() => {
      setLoadingRegions(false)
    })
  }, [getSharedRegions, getAvailableRegions])

  const getRegionName = useCallback(
    (regionId: string): string | undefined => {
      const region = [...availableRegions, ...sharedRegions].find((region) => region.id === regionId)
      return region?.name
    },
    [availableRegions, sharedRegions],
  )

  const customRegions = useMemo(() => {
    return availableRegions.filter((region) => region.regionType === RegionType.CUSTOM)
  }, [availableRegions])

  const contextValue: IRegionsContext = useMemo(() => {
    return {
      loadingRegions,
      sharedRegions,
      availableRegions,
      customRegions,
      refreshAvailableRegions: getAvailableRegions,
      getRegionName,
    }
  }, [loadingRegions, sharedRegions, availableRegions, customRegions, getAvailableRegions, getRegionName])

  return <RegionsContext.Provider value={contextValue}>{props.children}</RegionsContext.Provider>
}
