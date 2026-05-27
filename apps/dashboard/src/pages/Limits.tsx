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
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { UsageOverview, UsageOverviewSkeleton } from '@/components/UsageOverview'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { RoutePath } from '@/enums/RoutePath'
import { useOwnerTierQuery, useOwnerWalletQuery } from '@/hooks/queries/billingQueries'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useTiersQuery } from '@/hooks/queries/useTiersQuery'
import { useConfig } from '@/hooks/useConfig'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import type { Organization, RegionUsageOverview } from '@daytona/api-client'
import { keepPreviousData } from '@tanstack/react-query'
import { AlertCircle, ExternalLinkIcon, Globe, RefreshCcw, ShieldAlert } from 'lucide-react'
import { ReactNode, useEffect, useMemo, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate } from 'react-router-dom'

const PREVIEW_WARNING_CPU_QUOTA_THRESHOLD = 250
const NETWORK_LIMITS_DOCS_URL = `${DAYTONA_DOCS_URL}/en/network-limits/`
const PREVIEW_DOCS_URL = `${DAYTONA_DOCS_URL}/en/preview/`

export default function Limits() {
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const organizationTierQuery = useOwnerTierQuery()
  const walletQuery = useOwnerWalletQuery()
  const tiersQuery = useTiersQuery()

  const organizationTier = organizationTierQuery.data
  const tiers = tiersQuery.data?.slice().sort((a, b) => (a.tier ?? 0) - (b.tier ?? 0))
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
      <PageHeader />

      <PageContent>
        <PageIntro title="Limits" />
        {isError ? (
          <Card>
            <Empty className="py-12">
              <EmptyHeader>
                <EmptyMedia variant="icon" className="bg-destructive-background text-destructive">
                  <AlertCircle />
                </EmptyMedia>
                <EmptyTitle className="text-destructive">Failed to load limits</EmptyTitle>
                <EmptyDescription>Something went wrong while fetching limits data. Please try again.</EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button variant="secondary" size="sm" onClick={handleRetry}>
                  <RefreshCcw />
                  Retry
                </Button>
              </EmptyContent>
            </Empty>
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
                      value:
                        selectedOrganization?.sandboxLifecycleRateLimit || config?.rateLimit?.sandboxLifecycle?.limit,
                      label: 'Sandbox Lifecycle',
                      ttlSeconds:
                        selectedOrganization?.sandboxLifecycleRateLimitTtlSeconds ??
                        config?.rateLimit?.sandboxLifecycle?.ttl,
                    },
                  ]}
                />
                {selectedOrganization && (
                  <FeatureLimits organization={selectedOrganization} region={currentRegionUsageOverview} />
                )}
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
                        className="only:mb-4 border-l-0 border-r-0"
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

type ResourceType = 'compute' | 'memory' | 'storage'

interface LimitItem {
  value?: number | null
  unit?: string
  label: string
  ttlSeconds?: number | null
  resourceType?: ResourceType
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

function FeatureLimits({
  organization,
  region,
  className,
}: {
  organization: Organization
  region: RegionUsageOverview | null
  className?: string
}) {
  const previewWarningApplies = region != null && region.totalCpuQuota < PREVIEW_WARNING_CPU_QUOTA_THRESHOLD
  const networkEgressRestricted = organization.sandboxLimitedNetworkEgress

  return (
    <div className={cn('p-4 border-t border-border flex flex-col gap-4', className)}>
      <FeatureLimitItem
        icon={<ShieldAlert className="size-4" />}
        iconClassName="text-warning"
        label="Preview Warning"
        description={
          previewWarningApplies
            ? 'Shown before preview links open in a browser.'
            : 'Skipped for preview links in the selected region.'
        }
        docsHref={PREVIEW_DOCS_URL}
      />
      <FeatureLimitItem
        icon={<Globe className="size-4" />}
        iconClassName="text-green-500 dark:text-green-400"
        label="Internet Access"
        description={
          networkEgressRestricted
            ? 'Restricted egress. Essential development services remain available.'
            : 'Full outbound access is available by default.'
        }
        docsHref={NETWORK_LIMITS_DOCS_URL}
      />
    </div>
  )
}

function FeatureLimitItem({
  icon,
  iconClassName,
  label,
  description,
  docsHref,
}: {
  icon: ReactNode
  iconClassName?: string
  label: ReactNode
  description: ReactNode
  docsHref?: string
}) {
  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-3">
      <div className="flex min-w-40 items-center gap-3 text-sm font-medium text-foreground">
        <span className={cn('inline-flex size-5 shrink-0 items-center justify-center', iconClassName)}>{icon}</span>
        {label}
      </div>
      <div className="text-sm text-muted-foreground">
        {description}
        {docsHref ? (
          <>
            {' '}
            <a
              href={docsHref}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-foreground/80 underline underline-offset-4 hover:text-foreground"
            >
              Learn more
              <ExternalLinkIcon className="size-3.5" />
            </a>
          </>
        ) : null}
      </div>
    </div>
  )
}
