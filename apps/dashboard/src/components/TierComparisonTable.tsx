/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ComparisonSection, ComparisonTable } from './ComparisonTable'

import { OrganizationTier, Tier } from '@/billing-api'
import { TIER_RATE_LIMITS } from '@/constants/limits'
import { Skeleton } from './ui/skeleton'

export function TierComparisonTableSkeleton() {
  return (
    <div className="flex flex-col gap-4 p-4">
      {Array.from({ length: 5 }).map((_, index) => (
        <Skeleton key={index} className="h-7 w-full" />
      ))}
    </div>
  )
}

export function TierComparisonTable({
  tiers,
  currentTier,
  className,
}: {
  tiers: Tier[]
  currentTier?: OrganizationTier | null
  className?: string
}) {
  return (
    <ComparisonTable
      className={className}
      headerLabel="Tier"
      columns={[
        'Compute (vCPU)',
        'Memory (GiB)',
        'Storage (GiB)',
        'API Requests/min',
        'Sandbox Creation/min',
        'Sandbox Lifecycle/min',
      ]}
      currentRow={(currentTier?.tier || 1) - 1}
      data={buildTierComparisonTableData(tiers || [])}
    />
  )
}

function buildTierComparisonTableData(tiers: Tier[]): ComparisonSection[] {
  return [
    {
      id: 'tiers',
      title: 'Tiers',
      rows: tiers
        .map((tier) => {
          return {
            label: <span className="whitespace-nowrap">{tier.tier}</span>,
            values: [
              `${tier.tierLimit.concurrentCPU}`,
              `${tier.tierLimit.concurrentRAMGiB}`,
              `${tier.tierLimit.concurrentDiskGiB}`,
              `${TIER_RATE_LIMITS[tier.tier]?.authenticatedRateLimit.toLocaleString() || '-'}`,
              `${TIER_RATE_LIMITS[tier.tier]?.sandboxCreateRateLimit.toLocaleString() || '-'}`,
              `${TIER_RATE_LIMITS[tier.tier]?.sandboxLifecycleRateLimit.toLocaleString() || '-'}`,
            ],
          }
        })
        .concat({
          label: <span className="whitespace-nowrap">Enterprise</span>,
          values: Array(6).fill('Custom'),
        }),
    },
  ]
}
