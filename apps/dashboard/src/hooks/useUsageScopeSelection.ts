/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useMemo, useState } from 'react'
import { SandboxClass, type RegionUsageOverview } from '@daytona/api-client'

export interface UsageScopeSelection {
  classes: SandboxClass[]
  selectedSandboxClass: SandboxClass | undefined
  setSelectedSandboxClass: (sandboxClass: SandboxClass) => void
  regionsForSelectedClass: RegionUsageOverview[]
  selectedRegionId: string | undefined
  setSelectedRegionId: (regionId: string) => void
  currentEntry: RegionUsageOverview | null
}

export function useUsageScopeSelection(
  regionUsage: RegionUsageOverview[] | undefined,
  defaultRegionId?: string,
): UsageScopeSelection {
  const [selectedSandboxClass, setSelectedSandboxClass] = useState<SandboxClass | undefined>(undefined)
  const [selectedRegionId, setSelectedRegionId] = useState<string | undefined>(undefined)

  const usageByClass = useMemo(() => {
    const map = new Map<SandboxClass, RegionUsageOverview[]>()
    for (const usage of regionUsage ?? []) {
      const existing = map.get(usage.sandboxClass) ?? []
      existing.push(usage)
      map.set(usage.sandboxClass, existing)
    }
    return map
  }, [regionUsage])

  const classes = useMemo(() => Array.from(usageByClass.keys()), [usageByClass])

  useEffect(() => {
    if (classes.length === 0) {
      setSelectedSandboxClass(undefined)
      return
    }

    if (selectedSandboxClass && classes.includes(selectedSandboxClass)) {
      return
    }

    const defaultContainerClass =
      defaultRegionId &&
      usageByClass.get(SandboxClass.CONTAINER)?.some((usage) => usage.regionId === defaultRegionId) &&
      SandboxClass.CONTAINER
    const defaultRegionClass =
      defaultRegionId &&
      classes.find((sandboxClass) =>
        usageByClass.get(sandboxClass)?.some((usage) => usage.regionId === defaultRegionId),
      )
    const containerClass = classes.includes(SandboxClass.CONTAINER) ? SandboxClass.CONTAINER : undefined

    setSelectedSandboxClass(defaultContainerClass || defaultRegionClass || containerClass || classes[0])
  }, [classes, defaultRegionId, selectedSandboxClass, usageByClass])

  const regionsForSelectedClass = useMemo(() => {
    const regions = selectedSandboxClass ? (usageByClass.get(selectedSandboxClass) ?? []) : []

    return regions
      .map((usage, index) => ({ usage, index }))
      .sort((a, b) => {
        const aHasGpuQuota = a.usage.totalGpuQuota > 0
        const bHasGpuQuota = b.usage.totalGpuQuota > 0

        if (aHasGpuQuota !== bHasGpuQuota) {
          return bHasGpuQuota ? 1 : -1
        }

        return a.index - b.index
      })
      .map(({ usage }) => usage)
  }, [selectedSandboxClass, usageByClass])

  useEffect(() => {
    if (regionsForSelectedClass.length === 0) {
      setSelectedRegionId(undefined)
      return
    }

    if (selectedRegionId && regionsForSelectedClass.some((usage) => usage.regionId === selectedRegionId)) {
      return
    }

    const defaultRegion = defaultRegionId && regionsForSelectedClass.find((usage) => usage.regionId === defaultRegionId)
    setSelectedRegionId((defaultRegion || regionsForSelectedClass[0]).regionId)
  }, [defaultRegionId, regionsForSelectedClass, selectedRegionId])

  const currentEntry = useMemo(
    () => regionsForSelectedClass.find((usage) => usage.regionId === selectedRegionId) ?? null,
    [regionsForSelectedClass, selectedRegionId],
  )

  return {
    classes,
    selectedSandboxClass,
    setSelectedSandboxClass,
    regionsForSelectedClass,
    selectedRegionId,
    setSelectedRegionId,
    currentEntry,
  }
}
