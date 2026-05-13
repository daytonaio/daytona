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
  const { data: usageOverview } = useOrganizationUsageOverviewQuery({
    organizationId: selectedOrganization?.id ?? '',
  })

  return useMemo<SandboxClass[]>(() => {
    if (!regionId) return []
    const quotasForRegion = usageOverview?.regionUsage?.filter((r) => r.regionId === regionId) ?? []
    if (quotasForRegion.length > 0) {
      return [...new Set(quotasForRegion.map((q) => q.sandboxClass))]
    }
    return Object.values(SandboxClass)
  }, [usageOverview?.regionUsage, regionId])
}
