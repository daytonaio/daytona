/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMemo } from 'react'
import { SandboxClass } from '@daytona/api-client'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'

export function useAvailableSandboxClasses(regionId: string | undefined): SandboxClass[] {
  const { selectedOrganization } = useSelectedOrganization()
  const { data: usageOverview, isPending } = useOrganizationUsageOverviewQuery({
    organizationId: selectedOrganization?.id ?? '',
  })

  return useMemo<SandboxClass[]>(() => {
    if (!regionId) return []
    if (isPending || !usageOverview) return []
    const quotasForRegion = usageOverview.regionUsage?.filter((r) => r.regionId === regionId) ?? []
    if (quotasForRegion.length > 0) {
      return [...new Set(quotasForRegion.map((q) => q.sandboxClass))]
    }
    return Object.values(SandboxClass)
  }, [usageOverview, isPending, regionId])
}

export function useAvailableSandboxClassesForOrganization(): SandboxClass[] {
  const { selectedOrganization } = useSelectedOrganization()
  const { data: usageOverview, isPending } = useOrganizationUsageOverviewQuery({
    organizationId: selectedOrganization?.id ?? '',
  })

  return useMemo<SandboxClass[]>(() => {
    if (isPending || !usageOverview) return []
    const regionUsage = usageOverview.regionUsage ?? []
    if (regionUsage.length === 0) return Object.values(SandboxClass)
    return [...new Set(regionUsage.map((q) => q.sandboxClass))]
  }, [usageOverview, isPending])
}
