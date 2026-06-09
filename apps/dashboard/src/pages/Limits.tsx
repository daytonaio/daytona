/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { LiveIndicator } from '@/components/LiveIndicator'
import { PageContent, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { getSandboxClassLabel } from '@/components/SandboxTable/constants'
import { TierComparisonTable, TierComparisonTableSkeleton } from '@/components/TierComparisonTable'
import { TierUpgradeCard } from '@/components/TierUpgradeCard'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { UsageOverview, UsageOverviewSkeleton } from '@/components/UsageOverview'
import { RoutePath } from '@/enums/RoutePath'
import { useOwnerTierQuery, useOwnerWalletQuery } from '@/hooks/queries/billingQueries'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useTiersQuery } from '@/hooks/queries/useTiersQuery'
import { useConfig } from '@/hooks/useConfig'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useUsageScopeSelection } from '@/hooks/useUsageScopeSelection'
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
  const walletQuery = useOwnerWalletQuery()
  const tiersQuery = useTiersQuery()

  const organizationTier = organizationTierQuery.data
  const tiers = tiersQuery.data?.slice().sort((a, b) => (a.tier ?? 0) - (b.tier ?? 0))
  const wallet = walletQuery.data

  const { getRegionName } = useRegions()
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

  const {
    classes,
    selectedSandboxClass,
    setSelectedSandboxClass,
    regionsForSelectedClass,
    selectedRegionId,
    setSelectedRegionId,
    currentEntry: currentRegionUsageOverview,
  } = useUsageScopeSelection(usageOverview?.regionUsage, selectedOrganization?.defaultRegionId)
  const usageScopeAlerts = getUsageScopeAlerts(usageOverview?.regionUsage ?? [])

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
              <CardHeader className="p-4 space-y-0">
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
                </div>
                <CardDescription>
                  Limits help us mitigate misuse and manage infrastructure resources. <br /> Ensuring fair and stable
                  access to sandboxes and compute capacity across all users.
                </CardDescription>
                <UsageScopeAlertRow
                  alerts={usageScopeAlerts}
                  getRegionName={getRegionName}
                  onSelect={(alert) => {
                    setSelectedSandboxClass(alert.sandboxClass)
                    setSelectedRegionId(alert.regionId)
                  }}
                />
                {classes.length > 0 && (
                  <Tabs
                    value={selectedSandboxClass}
                    onValueChange={(value) => setSelectedSandboxClass(value as SandboxClass)}
                    className="-mx-4 gap-0 pt-5"
                  >
                    <div className="flex items-end justify-between gap-3 pr-4 shadow-[inset_0_-1px] shadow-border">
                      <div className="min-w-0 overflow-x-auto">
                        <TabsList variant="underline" className="min-w-max border-b-0">
                          {classes.map((sandboxClass) => {
                            const label = getSandboxClassLabel(sandboxClass)

                            return (
                              <TabsTrigger key={sandboxClass} value={sandboxClass} className="gap-2 px-4 py-2">
                                <span>{label}</span>
                              </TabsTrigger>
                            )
                          })}
                        </TabsList>
                      </div>
                      {regionsForSelectedClass.length > 0 && (
                        <div className="shrink-0 pb-1">
                          <Select
                            value={selectedRegionId}
                            onValueChange={setSelectedRegionId}
                            disabled={regionsForSelectedClass.length === 1}
                          >
                            <SelectTrigger
                              size="xs"
                              aria-label="Select region"
                              className={cn(
                                'w-auto min-w-20 max-w-40 gap-x-2 lowercase',
                                regionsForSelectedClass.length === 1 &&
                                  'pointer-events-none select-none disabled:opacity-100 [&>svg]:hidden',
                              )}
                            >
                              <SelectValue placeholder="Region" />
                            </SelectTrigger>
                            <SelectContent className="min-w-24 max-w-48" align="end">
                              {regionsForSelectedClass.map((usage) => (
                                <SelectItem key={usage.regionId} value={usage.regionId} className="lowercase">
                                  {getRegionName(usage.regionId) ?? usage.regionId}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      )}
                    </div>
                  </Tabs>
                )}
              </CardHeader>
              <CardContent className="p-0 flex flex-col">
                {usageOverviewQuery.isLoading ? (
                  <UsageOverviewSkeleton />
                ) : (
                  usageOverview &&
                  currentRegionUsageOverview && (
                    <div
                      className={cn('p-4 flex flex-col gap-2', {
                        'border-t border-border': classes.length === 0,
                      })}
                    >
                      <div className="flex items-center gap-4">
                        <div className="text-sm font-medium">Resources</div>
                        <LiveIndicator
                          isUpdating={usageOverviewQuery.isFetching}
                          intervalMs={10_000}
                          lastUpdatedAt={usageOverviewQuery.dataUpdatedAt || 0}
                        />
                      </div>
                      <UsageOverview
                        usageOverview={currentRegionUsageOverview}
                        hasGpuQuotaInClass={regionsForSelectedClass.some((usage) => usage.totalGpuQuota > 0)}
                      />
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

type UsageScopeSeverity = 'warning' | 'destructive'
type ResourceType = 'compute' | 'memory' | 'storage'
type UsageResourceLabel = 'CPU' | 'Memory' | 'Storage' | 'GPU'

interface UsageScopeAlert {
  key: string
  sandboxClass: SandboxClass
  regionId: string
  severity: UsageScopeSeverity
  resourceLabel: UsageResourceLabel
  percentage: number
}

interface LimitItem {
  value?: number | null
  unit?: string
  label: string
  ttlSeconds?: number | null
  resourceType?: ResourceType
}

function getUsageScopeAlerts(usages: RegionUsageOverview[]): UsageScopeAlert[] {
  return usages
    .map(getUsageScopeAlert)
    .filter((alert): alert is UsageScopeAlert => alert != null)
    .sort((a, b) => {
      if (a.severity !== b.severity) {
        return a.severity === 'destructive' ? -1 : 1
      }

      return b.percentage - a.percentage
    })
}

function getUsageScopeAlert(usage: RegionUsageOverview): UsageScopeAlert | null {
  const resources = [
    {
      label: 'CPU' as const,
      percentage: getUsagePercentage(usage.currentCpuUsage, usage.totalCpuQuota),
    },
    {
      label: 'Memory' as const,
      percentage: getUsagePercentage(usage.currentMemoryUsage, usage.totalMemoryQuota),
    },
    {
      label: 'Storage' as const,
      percentage: getUsagePercentage(usage.currentDiskUsage, usage.totalDiskQuota),
    },
    {
      label: 'GPU' as const,
      percentage: getUsagePercentage(usage.currentGpuUsage, usage.totalGpuQuota),
    },
  ].filter((resource): resource is { label: UsageResourceLabel; percentage: number } => resource.percentage != null)

  const highestUsageResource = resources.reduce<(typeof resources)[number] | null>((highest, resource) => {
    if (!highest || resource.percentage > highest.percentage) {
      return resource
    }

    return highest
  }, null)

  if (!highestUsageResource || highestUsageResource.percentage <= 60) {
    return null
  }

  return {
    key: `${usage.sandboxClass}-${usage.regionId}`,
    sandboxClass: usage.sandboxClass,
    regionId: usage.regionId,
    severity: highestUsageResource.percentage > 90 ? 'destructive' : 'warning',
    resourceLabel: highestUsageResource.label,
    percentage: highestUsageResource.percentage,
  }
}

function getUsagePercentage(current: number, total: number): number | null {
  if (total <= 0) return null
  return (current / total) * 100
}

function UsageScopeAlertRow({
  alerts,
  getRegionName,
  onSelect,
}: {
  alerts: UsageScopeAlert[]
  getRegionName: (regionId: string) => string | undefined
  onSelect: (alert: UsageScopeAlert) => void
}) {
  if (alerts.length === 0) return null

  return (
    <div className="flex min-w-0 gap-1.5 overflow-x-auto pt-4 text-xs">
      {alerts.map((alert) => (
        <button
          key={alert.key}
          type="button"
          className={cn(
            'inline-flex h-6 shrink-0 items-center gap-1 rounded-full border px-2 text-xs transition-colors',
            alert.severity === 'destructive'
              ? 'border-destructive-separator bg-destructive-background text-destructive-foreground hover:bg-destructive-background/80'
              : 'border-warning-separator bg-warning-background text-warning-foreground hover:bg-warning-background/80',
          )}
          onClick={() => onSelect(alert)}
        >
          <span className="font-medium">
            {getSandboxClassLabel(alert.sandboxClass)} - {getRegionName(alert.regionId) ?? alert.regionId}
          </span>
          <span className="opacity-70">
            {alert.resourceLabel} {Math.round(alert.percentage)}%
          </span>
        </button>
      ))}
    </div>
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
