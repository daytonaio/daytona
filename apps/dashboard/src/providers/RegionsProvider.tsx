/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ReactNode, useCallback, useEffect, useMemo, useState } from 'react'
import { Region } from '@daytonaio/api-client'
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

  const getRegions = useCallback(async () => {
    if (!selectedOrganization) {
      setRegions([])
      return []
    }
    try {
      const regions = (await regionsApi.listRegions(selectedOrganization.id)).data
      setRegions(regions)
      return regions
    } catch (error) {
      handleApiError(error, 'Failed to fetch regions')
      throw error
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

  const contextValue: IRegionsContext = useMemo(() => {
    return {
      regions,
      refreshRegions: getRegions,
      getRegionName,
    }
  }, [regions, getRegions, getRegionName])

  return <RegionsContext.Provider value={contextValue}>{props.children}</RegionsContext.Provider>
}
