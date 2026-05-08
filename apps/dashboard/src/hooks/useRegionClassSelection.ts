/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useMemo, useState } from 'react'
import { RegionUsageOverview, SandboxClass } from '@daytona/api-client'

export interface RegionClassSelection {
  regionIds: string[]
  selectedRegionId: string | undefined
  setSelectedRegionId: (id: string) => void
  classesForSelectedRegion: RegionUsageOverview[]
  selectedSandboxClass: SandboxClass | undefined
  setSelectedSandboxClass: (cls: SandboxClass) => void
  showClassSelector: boolean
  currentEntry: RegionUsageOverview | null
}

export function useRegionClassSelection(
  regionUsage: RegionUsageOverview[] | undefined,
  defaultRegionId?: string,
): RegionClassSelection {
  const [selectedRegionId, setSelectedRegionId] = useState<string | undefined>(undefined)
  const [selectedSandboxClass, setSelectedSandboxClass] = useState<SandboxClass | undefined>(undefined)

  const groupedByRegion = useMemo(() => {
    const map = new Map<string, RegionUsageOverview[]>()
    if (!regionUsage) return map
    for (const usage of regionUsage) {
      const list = map.get(usage.regionId) ?? []
      list.push(usage)
      map.set(usage.regionId, list)
    }
    return map
  }, [regionUsage])

  const regionIds = useMemo(() => Array.from(groupedByRegion.keys()), [groupedByRegion])

  useEffect(() => {
    if (selectedRegionId || regionIds.length === 0) return
    const initial = defaultRegionId && regionIds.includes(defaultRegionId) ? defaultRegionId : regionIds[0]
    setSelectedRegionId(initial)
  }, [regionIds, defaultRegionId, selectedRegionId])

  const classesForSelectedRegion = useMemo(
    () => (selectedRegionId ? (groupedByRegion.get(selectedRegionId) ?? []) : []),
    [groupedByRegion, selectedRegionId],
  )

  useEffect(() => {
    if (classesForSelectedRegion.length === 0) return
    if (selectedSandboxClass && classesForSelectedRegion.some((c) => c.sandboxClass === selectedSandboxClass)) {
      return
    }
    const preferred =
      classesForSelectedRegion.find((c) => c.sandboxClass === SandboxClass.CONTAINER) ?? classesForSelectedRegion[0]
    setSelectedSandboxClass(preferred.sandboxClass)
  }, [classesForSelectedRegion, selectedSandboxClass])

  const currentEntry = useMemo(() => {
    if (!selectedSandboxClass) return null
    return classesForSelectedRegion.find((c) => c.sandboxClass === selectedSandboxClass) ?? null
  }, [classesForSelectedRegion, selectedSandboxClass])

  return {
    regionIds,
    selectedRegionId,
    setSelectedRegionId,
    classesForSelectedRegion,
    selectedSandboxClass,
    setSelectedSandboxClass,
    showClassSelector: classesForSelectedRegion.length > 1,
    currentEntry,
  }
}
