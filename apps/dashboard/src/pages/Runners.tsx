/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { RegionDto, Runner, OrganizationRolePermissionsEnum } from '@daytonaio/api-client'
import { RunnerTable } from '@/components/RunnerTable'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { handleApiError } from '@/lib/error-handling'
import { Server } from 'lucide-react'

const Runners: React.FC = () => {
  const { regionsApi, runnersApi } = useApi()
  const [regions, setRegions] = useState<RegionDto[]>([])
  const [runners, setRunners] = useState<Runner[]>([])
  const [loadingRegions, setLoadingRegions] = useState(true)
  const [loadingRunners, setLoadingRunners] = useState(false)
  const [selectedRegion, setSelectedRegion] = useState<string>('')
  const [loadingTable, setLoadingTable] = useState(false)

  const { selectedOrganization, authenticatedUserHasPermission } = useSelectedOrganization()

  const fetchRegions = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setLoadingRegions(true)
    try {
      const response = (await regionsApi.listRegions(selectedOrganization.id)).data
      setRegions(response)
      // Auto-select the first region if available
      if (response.length > 0 && !selectedRegion) {
        setSelectedRegion(response[0].name)
      }
    } catch (error) {
      handleApiError(error, 'Failed to fetch regions')
    } finally {
      setLoadingRegions(false)
    }
  }, [regionsApi, selectedOrganization, selectedRegion])

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_REGIONS),
    [authenticatedUserHasPermission],
  )

  const fetchRunners = useCallback(async () => {
    if (!selectedRegion) {
      return
    }
    setLoadingRunners(true)
    setLoadingTable(true)
    try {
      // TODO: Fix when listRunners returns proper data
      // const response = (await runnersApi.listRunners(selectedRegion)).data
      // setRunners(response || [])

      // Mock response for now - remove this when API is fixed
      setRunners([
        {
          id: 'mock-runner-1',
          domain: 'runner1.example.com',
          apiUrl: 'https://api.runner1.example.com',
          proxyUrl: 'https://proxy.runner1.example.com',
          apiKey: 'mock-key',
          cpu: 8,
          memory: 16,
          disk: 100,
          gpu: 1,
          gpuType: 'RTX 4090',
          class: 'large',
          used: 2,
          capacity: 10,
          currentCpuUsagePercentage: 45.6,
          currentMemoryUsagePercentage: 68.2,
          currentDiskUsagePercentage: 33.8,
          currentAllocatedCpu: 4000,
          currentAllocatedMemoryGiB: 8000,
          currentAllocatedDiskGiB: 50000,
          currentSnapshotCount: 12,
          availabilityScore: 85,
          region: selectedRegion,
          state: 'ready',
          version: '1.0.0',
          lastChecked: new Date().toISOString(),
          unschedulable: false,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        } as any,
      ])
    } catch (error) {
      handleApiError(error, 'Failed to fetch runners')
      setRunners([])
    } finally {
      setLoadingRunners(false)
      setLoadingTable(false)
    }
  }, [runnersApi, selectedRegion])

  useEffect(() => {
    fetchRegions()
  }, [fetchRegions])

  useEffect(() => {
    if (selectedRegion) {
      fetchRunners()
    }
  }, [fetchRunners])

  const handleRegionChange = (regionName: string) => {
    setSelectedRegion(regionName)
  }

  if (loadingRegions) {
    return (
      <div className="px-6 py-2">
        <div className="mb-2 h-12 flex items-center justify-between">
          <h1 className="text-2xl font-medium">Runners</h1>
        </div>
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-12 bg-muted animate-pulse rounded" />
          ))}
        </div>
      </div>
    )
  }

  if (regions.length === 0) {
    return (
      <div className="px-6 py-2">
        <div className="mb-2 h-12 flex items-center justify-between">
          <h1 className="text-2xl font-medium">Runners</h1>
        </div>
        <div className="text-center py-12">
          <Server className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No regions found.</p>
          <p className="text-sm text-muted-foreground mt-1">Create a region first to view runners.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Runners</h1>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Region:</span>
            <Select value={selectedRegion} onValueChange={handleRegionChange}>
              <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="Select a region" />
              </SelectTrigger>
              <SelectContent>
                {regions.map((region) => (
                  <SelectItem key={region.name} value={region.name}>
                    {region.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {selectedRegion && <RunnerTable data={runners} loading={loadingTable} writePermitted={writePermitted} />}
    </div>
  )
}

export default Runners
