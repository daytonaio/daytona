/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { LiveIndicator } from '@/components/LiveIndicator'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { TierComparisonTable, TierComparisonTableSkeleton } from '@/components/TierComparisonTable'
import { TierUpgradeCard } from '@/components/TierUpgradeCard'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { UsageOverview, UsageOverviewSkeleton } from '@/components/UsageOverview'
import { RoutePath } from '@/enums/RoutePath'
import { useOwnerTierQuery, useOwnerWalletQuery } from '@/hooks/queries/billingQueries'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useTiersQuery } from '@/hooks/queries/useTiersQuery'
import { useConfig } from '@/hooks/useConfig'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import type { RegionUsageOverview } from '@daytonaio/api-client'
import { keepPreviousData } from '@tanstack/react-query'
import { RefreshCcw } from 'lucide-react'
import { ReactNode, useEffect, useMemo, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate } from 'react-router-dom'

export default function Limits() {
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const organizationTierQuery = useOwnerTierQuery()
  const walletQuery = useOwnerWalletQuery()
  const tiersQuery = useTiersQuery()

  const organizationTier = organizationTierQuery.data
  const tiers = tiersQuery.data
  const wallet = walletQuery.data

  const { getRegionName } = useRegions()
  const [selectedRegionId, setSelectedRegionId] = useState<string | undefined>(undefined)
  const config = useConfig()
  const navigate = useNavigate()

  useEffect(() => {
    if (selectedOrganization && !selectedOrganization.defaultRegionId) {
      navigate(RoutePath.SETTINGS)
    }
  }, [navigate, selectedOrganization])

  const { data: usageOverview, ...usageOverviewQuery } = useOrganizationUsageOverviewQuery(
    {
      organizationId: selectedOrganization?.id || '',
    },
    {
      placeholderData: keepPreviousData,
      refetchInterval: 10_000,
      refetchIntervalInBackground: true,
      staleTime: 0,
    },
  )

  useEffect(() => {
    if (usageOverview && usageOverview.regionUsage.length > 0 && !selectedRegionId) {
      const regionIds = usageOverview.regionUsage.map((usage) => usage.regionId)
      const regionId = regionIds.find((regionId) => regionId === selectedOrganization?.defaultRegionId) || regionIds[0]
      setSelectedRegionId(regionId)
    }
  }, [usageOverview, selectedOrganization?.defaultRegionId, selectedRegionId])

  const currentRegionUsageOverview = useMemo<RegionUsageOverview | null>(() => {
    if (!usageOverview || !selectedRegionId) {
      return null
    }
    return usageOverview.regionUsage.find((usage) => usage.regionId === selectedRegionId) || null
  }, [usageOverview, selectedRegionId])

  const isLoading = organizationTierQuery.isLoading || tiersQuery.isLoading || walletQuery.isLoading
  const isError =
    organizationTierQuery.isError || tiersQuery.isError || usageOverviewQuery.isError || walletQuery.isError

  const handleRetry = () => {
    organizationTierQuery.refetch()
    tiersQuery.refetch()
    usageOverviewQuery.refetch()
    walletQuery.refetch()
  }

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Limits</PageTitle>
      </PageHeader>

      <PageContent>
        {isError ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-center">Oops, something went wrong</CardTitle>
            </CardHeader>
            <CardContent className="flex justify-between items-center flex-col gap-3">
              <div>There was an error loading your limits.</div>
              <Button variant="outline" onClick={handleRetry}>
                <RefreshCcw className="mr-2 h-4 w-4" />
                Retry
              </Button>
            </CardContent>
          </Card>
        ) : (
          <>
            <Card>
              <CardHeader className="p-4">
                <div className="flex items-center justify-between gap-2 mb-2 flex-wrap">
                  <CardTitle className="flex justify-between gap-x-4 gap-y-2 flex-row flex-wrap items-center">
                    <div className="flex items-center gap-2">
                      Current Usage{' '}
                      {organizationTier && (
                        <Badge variant="outline" className="font-mono uppercase">
                          Tier {organizationTier.tier}
                        </Badge>
                      )}
                    </div>
                  </CardTitle>
                  {usageOverview && usageOverview.regionUsage.length > 0 && (
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-muted-foreground">Region:</span>
                      <Select value={selectedRegionId} onValueChange={setSelectedRegionId}>
                        <SelectTrigger
                          size="xs"
                          disabled={usageOverview.regionUsage.length === 1}
                          className={`uppercase w-auto min-w-12 max-w-48 gap-x-2 ${usageOverview.regionUsage.length === 1 ? 'pointer-events-none select-none [&>svg]:hidden min-w-10 disabled:opacity-100' : ''}`}
                        >
                          <SelectValue placeholder="Select region" />
                        </SelectTrigger>
                        <SelectContent className="min-w-24 max-w-48" align="end">
                          {usageOverview.regionUsage.map((usage) => (
                            <SelectItem key={usage.regionId} value={usage.regionId} className="uppercase">
                              {getRegionName(usage.regionId) ?? usage.regionId}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>
                <CardDescription>
                  Limits help us mitigate misuse and manage infrastructure resources. <br /> Ensuring fair and stable
                  access to sandboxes and compute capacity across all users.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-0 flex flex-col">
                {usageOverviewQuery.isLoading ? (
                  <UsageOverviewSkeleton />
                ) : (
                  usageOverview &&
                  currentRegionUsageOverview && (
                    <div className="p-4 border-t border-border flex flex-col gap-2">
                      <div className="flex items-center gap-4">
                        <div className="text-sm font-medium">Resources</div>
                        <LiveIndicator
                          isUpdating={usageOverviewQuery.isFetching}
                          intervalMs={10_000}
                          lastUpdatedAt={usageOverviewQuery.dataUpdatedAt || 0}
                        />
                      </div>
                      <UsageOverview usageOverview={currentRegionUsageOverview} />
                    </div>
                  )
                )}
                <RateLimits
                  title="Sandbox Limits"
                  description="Resources limit per sandbox."
                  className="border-t border-border"
                  rateLimits={[
                    { label: 'Compute', value: selectedOrganization?.maxCpuPerSandbox, unit: 'vCPU' },
                    { label: 'Memory', value: selectedOrganization?.maxMemoryPerSandbox, unit: 'GiB' },
                    { label: 'Storage', value: selectedOrganization?.maxDiskPerSandbox, unit: 'GiB' },
                  ]}
                />

                <RateLimits
                  title="Rate Limits"
                  description="How many requests you can make."
                  className="border-t border-border"
                  rateLimits={[
                    {
                      value: selectedOrganization?.authenticatedRateLimit || config?.rateLimit?.authenticated?.limit,
                      label: 'General Requests',
                      ttlSeconds:
                        selectedOrganization?.authenticatedRateLimitTtlSeconds ?? config?.rateLimit?.authenticated?.ttl,
                    },
                    {
                      value: selectedOrganization?.sandboxCreateRateLimit || config?.rateLimit?.sandboxCreate?.limit,
                      label: 'Sandbox Creation',
                      ttlSeconds:
                        selectedOrganization?.sandboxCreateRateLimitTtlSeconds ?? config?.rateLimit?.sandboxCreate?.ttl,
                    },
                    {
                      value:
                        selectedOrganization?.sandboxLifecycleRateLimit || config?.rateLimit?.sandboxLifecycle?.limit,
                      label: 'Sandbox Lifecycle',
                      ttlSeconds:
                        selectedOrganization?.sandboxLifecycleRateLimitTtlSeconds ??
                        config?.rateLimit?.sandboxLifecycle?.ttl,
                    },
                  ]}
                />
              </CardContent>
            </Card>

            {config.billingApiUrl && selectedOrganization && (
              <>
                <TierUpgradeCard
                  organizationTier={organizationTier}
                  tiers={tiers || []}
                  organization={selectedOrganization}
                  requirementsState={{
                    emailVerified: !!user?.profile?.email_verified,
                    creditCardLinked: !!wallet?.creditCardConnected,
                  }}
                />

                <Card className="mb-10">
                  <CardHeader>
                    <CardTitle className="flex items-center mb-2">Limits</CardTitle>
                  </CardHeader>
                  <CardContent className="p-0">
                    {isLoading ? (
                      <TierComparisonTableSkeleton />
                    ) : (
                      <TierComparisonTable
                        className="border-l-0 border-r-0 rounded-none only:mb-4"
                        tiers={tiers || []}
                        currentTier={organizationTier}
                      />
                    )}
                  </CardContent>
                </Card>
              </>
            )}
          </>
        )}
      </PageContent>
    </PageLayout>
  )
}

interface LimitItem {
  value?: number | null
  unit?: string
  label: string
  ttlSeconds?: number | null
}

function RateLimits({
  rateLimits,
  className,
  title,
  description,
}: {
  rateLimits: LimitItem[]
  className?: string
  title: ReactNode
  description: ReactNode
}) {
  const isEmpty = rateLimits.every(({ value }) => !value)
  if (isEmpty) {
    return null
  }

  return (
    <div className={cn('p-4 border-t border-border flex flex-col gap-4', className)}>
      <div className="flex flex-col gap-1">
        <div className="text-foreground text-sm font-medium">{title}</div>
        <div className="text-muted-foreground text-sm">{description}</div>
      </div>
      <div className="grid grid-cols-1 gap-2 sm:gap-4 sm:grid-cols-3">
        {rateLimits.map(
          ({ label, value, unit, ttlSeconds }) =>
            value && <RateLimitItem key={label} label={label} value={value} unit={unit} ttlSeconds={ttlSeconds} />,
        )}
      </div>
    </div>
  )
}

function formatTtl(ttlSeconds?: number | null): string {
  if (!ttlSeconds) return ' / min'
  if (ttlSeconds % 60 === 0) return ` / ${ttlSeconds / 60}min`
  return ` / ${ttlSeconds}s`
}

function RateLimitItem({ label, value, unit, ttlSeconds }: LimitItem) {
  if (!value) {
    return null
  }

  return (
    <div className="flex flex-col">
      <div className="text-muted-foreground text-xs">{label}</div>
      <div className="text-foreground text-sm font-medium">
        {value?.toLocaleString()}
        {unit ? ` ${unit}` : formatTtl(ttlSeconds)}
      </div>
    </div>
  )
}
