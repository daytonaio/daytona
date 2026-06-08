/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { LiveIndicator } from '@/components/LiveIndicator'
import { PageContent, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { TierComparisonTable, TierComparisonTableSkeleton } from '@/components/TierComparisonTable'
import { TierUpgradeCard } from '@/components/TierUpgradeCard'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { UsageOverview, UsageOverviewSkeleton } from '@/components/UsageOverview'
import { RoutePath } from '@/enums/RoutePath'
import { useOwnerTierQuery } from '@/hooks/queries/billingQueries'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { usePaymentMethodsQuery } from '@/hooks/queries/usePaymentMethodsQuery'
import { useTiersQuery } from '@/hooks/queries/useTiersQuery'
import { useConfig } from '@/hooks/useConfig'
import { useRegions } from '@/hooks/useRegions'
import { useRegionClassSelection } from '@/hooks/useRegionClassSelection'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import type { Organization, RegionUsageOverview, SandboxClass } from '@daytona/api-client'
import { keepPreviousData } from '@tanstack/react-query'
import { RefreshCcw } from 'lucide-react'
import { ReactNode, useEffect } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate } from 'react-router-dom'

export default function Limits() {
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const organizationTierQuery = useOwnerTierQuery()
  const tiersQuery = useTiersQuery()

  const organizationTier = organizationTierQuery.data
  const tiers = tiersQuery.data?.slice().sort((a, b) => (a.tier ?? 0) - (b.tier ?? 0))

  const { getRegionName } = useRegions()
  const config = useConfig()
  const navigate = useNavigate()
  const paymentMethodsQuery = usePaymentMethodsQuery({
    organizationId: selectedOrganization?.id ?? '',
    enabled: Boolean(config.billingApiUrl && selectedOrganization),
  })
  const paymentMethods = paymentMethodsQuery.data
  const paymentMethodsUnavailable = paymentMethodsQuery.isError && paymentMethods === undefined
  const hasPaymentMethod = (paymentMethods?.length ?? 0) > 0

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

  const {
    regionIds,
    selectedRegionId,
    setSelectedRegionId,
    classesForSelectedRegion,
    selectedSandboxClass,
    setSelectedSandboxClass,
    showClassSelector,
    currentEntry: currentRegionUsageOverview,
  } = useRegionClassSelection(usageOverview?.regionUsage, selectedOrganization?.defaultRegionId)

  const tierDataLoading = organizationTierQuery.isLoading || tiersQuery.isLoading
  const tierDataError = organizationTierQuery.isError || tiersQuery.isError
  const upgradeRequirementsLoading = tierDataLoading || paymentMethodsQuery.isLoading
  const upgradeRequirementsError = paymentMethodsUnavailable && !tierDataError

  const handleRetryUsage = () => {
    usageOverviewQuery.refetch()
  }

  const handleRetryTierData = () => {
    organizationTierQuery.refetch()
    tiersQuery.refetch()
  }

  const handleRetryPaymentMethods = () => {
    paymentMethodsQuery.refetch()
  }

  return (
    <PageLayout>
      <PageHeader />

      <PageContent>
        <PageIntro title="Limits" />
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
              {regionIds.length > 0 && (
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">Region:</span>
                  <Select value={selectedRegionId} onValueChange={setSelectedRegionId}>
                    <SelectTrigger
                      size="xs"
                      disabled={regionIds.length === 1}
                      className={`uppercase w-auto min-w-12 max-w-48 gap-x-2 ${regionIds.length === 1 ? 'pointer-events-none select-none [&>svg]:hidden min-w-10 disabled:opacity-100' : ''}`}
                    >
                      <SelectValue placeholder="Select region" />
                    </SelectTrigger>
                    <SelectContent className="min-w-24 max-w-48" align="end">
                      {regionIds.map((regionId) => (
                        <SelectItem key={regionId} value={regionId} className="uppercase">
                          {getRegionName(regionId) ?? regionId}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {showClassSelector && (
                    <>
                      <span className="text-sm text-muted-foreground">Class:</span>
                      <Select
                        value={selectedSandboxClass}
                        onValueChange={(value) => setSelectedSandboxClass(value as SandboxClass)}
                      >
                        <SelectTrigger size="xs" className="uppercase w-auto min-w-12 max-w-48 gap-x-2">
                          <SelectValue placeholder="Select class" />
                        </SelectTrigger>
                        <SelectContent className="min-w-24 max-w-48" align="end">
                          {classesForSelectedRegion.map((usage) => (
                            <SelectItem key={usage.sandboxClass} value={usage.sandboxClass} className="uppercase">
                              {usage.sandboxClass}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </>
                  )}
                </div>
              )}
            </div>
            <CardDescription>
              Limits help us mitigate misuse and manage infrastructure resources. <br /> Ensuring fair and stable access
              to sandboxes and compute capacity across all users.
            </CardDescription>
          </CardHeader>
          <CardContent className="p-0 flex flex-col">
            {usageOverviewQuery.isLoading ? (
              <UsageOverviewSkeleton />
            ) : usageOverviewQuery.isError ? (
              <div className="border-t border-border flex">
                <LimitsSectionErrorState
                  title="Failed to load resource usage"
                  description="Something went wrong while fetching your current resource usage."
                  onRetry={handleRetryUsage}
                />
              </div>
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
              rateLimits={buildSandboxLimitItems(currentRegionUsageOverview, selectedOrganization)}
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
                  value: selectedOrganization?.sandboxLifecycleRateLimit || config?.rateLimit?.sandboxLifecycle?.limit,
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
            {upgradeRequirementsError ? (
              <Card>
                <CardContent className="flex p-0">
                  <LimitsSectionErrorState
                    title="Failed to load upgrade requirements"
                    description="Something went wrong while fetching your billing requirements."
                    onRetry={handleRetryPaymentMethods}
                  />
                </CardContent>
              </Card>
            ) : (
              !upgradeRequirementsLoading &&
              !tierDataError && (
                <TierUpgradeCard
                  organizationTier={organizationTier}
                  tiers={tiers || []}
                  organization={selectedOrganization}
                  requirementsState={{
                    emailVerified: !!user?.profile?.email_verified,
                    creditCardLinked: hasPaymentMethod,
                  }}
                />
              )
            )}

            <Card className="mb-10">
              <CardHeader>
                <CardTitle className="flex items-center mb-2">Limits</CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                {tierDataLoading ? (
                  <TierComparisonTableSkeleton />
                ) : tierDataError ? (
                  <div className="flex">
                    <LimitsSectionErrorState
                      title="Failed to load tier limits"
                      description="Something went wrong while fetching billing tiers."
                      onRetry={handleRetryTierData}
                    />
                  </div>
                ) : (
                  <TierComparisonTable
                    className="only:mb-4 border-l-0 border-r-0"
                    tiers={tiers || []}
                    currentTier={organizationTier}
                  />
                )}
              </CardContent>
            </Card>
          </>
        )}
      </PageContent>
    </PageLayout>
  )
}

type ResourceType = 'compute' | 'memory' | 'storage'

interface LimitItem {
  value?: number | null
  unit?: string
  label: string
  ttlSeconds?: number | null
  resourceType?: ResourceType
}

function LimitsSectionErrorState({
  title,
  description,
  onRetry,
}: {
  title: string
  description: string
  onRetry: () => void
}) {
  return (
    <Empty className="flex-1 border-0 py-8">
      <EmptyHeader>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription>{description}</EmptyDescription>
      </EmptyHeader>
      <Button variant="outline" size="sm" onClick={onRetry}>
        <RefreshCcw className="size-4" />
        Retry
      </Button>
    </Empty>
  )
}

function buildSandboxLimitItems(region: RegionUsageOverview | null, org: Organization | null | undefined): LimitItem[] {
  const items: LimitItem[] = []
  const gpuEnabled = (region?.totalGpuQuota ?? 0) > 0

  const cpuBase = region?.maxCpuPerSandbox ?? org?.maxCpuPerSandbox
  const cpuGpu = gpuEnabled ? region?.maxCpuPerGpuSandbox : null
  items.push({ resourceType: 'compute', label: 'Compute', value: cpuBase, unit: 'vCPU' })
  if (cpuGpu != null && cpuGpu !== cpuBase) {
    items.push({ resourceType: 'compute', label: 'Compute (GPU)', value: cpuGpu, unit: 'vCPU' })
  }

  const memBase = region?.maxMemoryPerSandbox ?? org?.maxMemoryPerSandbox
  const memGpu = gpuEnabled ? region?.maxMemoryPerGpuSandbox : null
  items.push({ resourceType: 'memory', label: 'Memory', value: memBase, unit: 'GiB' })
  if (memGpu != null && memGpu !== memBase) {
    items.push({ resourceType: 'memory', label: 'Memory (GPU)', value: memGpu, unit: 'GiB' })
  }

  const diskBase = region?.maxDiskPerSandbox ?? org?.maxDiskPerSandbox
  const diskNonEphem = region?.maxDiskPerNonEphemeralSandbox
  const diskGpu = gpuEnabled ? region?.maxDiskPerGpuSandbox : null

  const showNonEphemSplit = diskNonEphem != null && diskNonEphem > 0 && diskNonEphem !== diskBase
  const showStorageGpuVariant = diskGpu != null && diskGpu !== diskBase

  items.push({
    resourceType: 'storage',
    label: showNonEphemSplit ? 'Storage (Ephemeral)' : 'Storage',
    value: diskBase,
    unit: 'GiB',
  })

  if (showStorageGpuVariant) {
    items.push({ resourceType: 'storage', label: 'Storage (GPU)', value: diskGpu, unit: 'GiB' })
  }

  if (showNonEphemSplit) {
    items.push({ resourceType: 'storage', label: 'Storage (Non-Ephemeral)', value: diskNonEphem, unit: 'GiB' })
  }

  return items
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
  if (isEmpty) return null

  // When items carry resourceType, render each type as its own desktop column
  // (variants stack vertically within the column). Otherwise fall back to the
  // flat grid used by sections like "Rate Limits".
  const grouped = groupByResourceType(rateLimits)

  return (
    <div className={cn('p-4 border-t border-border flex flex-col gap-4', className)}>
      <div className="flex flex-col gap-1">
        <div className="text-foreground text-sm font-medium">{title}</div>
        <div className="text-muted-foreground text-sm">{description}</div>
      </div>
      {grouped ? (
        <div className="grid grid-cols-1 gap-2 sm:gap-4 sm:grid-cols-3">
          {grouped.map((column) => (
            <div key={column[0].resourceType} className="flex flex-col gap-2 sm:gap-4">
              {column.map(
                ({ label, value, unit, ttlSeconds }) =>
                  value && (
                    <RateLimitItem key={label} label={label} value={value} unit={unit} ttlSeconds={ttlSeconds} />
                  ),
              )}
            </div>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-2 sm:gap-4 sm:grid-cols-3">
          {rateLimits.map(
            ({ label, value, unit, ttlSeconds }) =>
              value && <RateLimitItem key={label} label={label} value={value} unit={unit} ttlSeconds={ttlSeconds} />,
          )}
        </div>
      )}
    </div>
  )
}

function groupByResourceType(items: LimitItem[]): LimitItem[][] | null {
  if (!items.some((item) => item.resourceType !== undefined)) {
    return null
  }
  const groups = new Map<ResourceType, LimitItem[]>()
  for (const item of items) {
    if (!item.resourceType) continue
    const existing = groups.get(item.resourceType)
    if (existing) {
      existing.push(item)
    } else {
      groups.set(item.resourceType, [item])
    }
  }
  return Array.from(groups.values())
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
