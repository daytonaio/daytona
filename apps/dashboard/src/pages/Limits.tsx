/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { RegionUsageOverview } from '@daytonaio/api-client'
import { LiveIndicator } from '@/components/LiveIndicator'
import { TierComparisonTable, TierComparisonTableSkeleton } from '@/components/TierComparisonTable'
import { TierUpgradeCard } from '@/components/TierUpgradeCard'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { UsageOverview, UsageOverviewSkeleton } from '@/components/UsageOverview'
import { RoutePath } from '@/enums/RoutePath'
import { useOrganizationTierQuery } from '@/hooks/queries/useOrganizationTierQuery'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useOrganizationWalletQuery } from '@/hooks/queries/useOrganizationWalletQuery'
import { useTiersQuery } from '@/hooks/queries/useTiersQuery'
import { useConfig } from '@/hooks/useConfig'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import { keepPreviousData } from '@tanstack/react-query'
import { RefreshCcw } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate } from 'react-router-dom'
import { UserProfileIdentity } from './LinkedAccounts'

export default function Limits() {
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const { getRegionName } = useRegions()
  const [selectedRegionId, setSelectedRegionId] = useState<string | undefined>(undefined)
  const config = useConfig()
  const navigate = useNavigate()

  useEffect(() => {
    if (selectedOrganization && !selectedOrganization.defaultRegionId) {
      navigate(RoutePath.SETTINGS)
    }
  }, [navigate, selectedOrganization])

  const { data: organizationTier, ...organizationTierQuery } = useOrganizationTierQuery({
    organizationId: selectedOrganization?.id || '',
  })
  const { data: wallet, ...walletQuery } = useOrganizationWalletQuery({
    organizationId: selectedOrganization?.id || '',
  })
  const { data: tiers, ...tiersQuery } = useTiersQuery()
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
      setSelectedRegionId(usageOverview.regionUsage[0].regionId)
    }
  }, [usageOverview, selectedRegionId])

  const githubConnected = useMemo(() => {
    if (!user?.profile?.identities) {
      return false
    }
    return (user.profile.identities as UserProfileIdentity[]).some(
      (identity: UserProfileIdentity) => identity.provider === 'github',
    )
  }, [user])

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
    <div className="px-6 py-2 max-w-3xl p-5">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Limits</h1>
      </div>

      <div className="flex flex-col gap-6">
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
                          className={`w-auto min-w-12 max-w-48 gap-x-2 ${usageOverview.regionUsage.length === 1 ? 'pointer-events-none select-none [&>svg]:hidden min-w-10' : ''}`}
                        >
                          <SelectValue placeholder="Select region" />
                        </SelectTrigger>
                        <SelectContent className="min-w-24 max-w-48" align="end">
                          {usageOverview.regionUsage.map((usage) => (
                            <SelectItem key={usage.regionId} value={usage.regionId}>
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
                  className="border-t border-border"
                  rateLimits={[
                    {
                      value: config?.rateLimit?.authenticated?.limit || selectedOrganization?.authenticatedRateLimit,
                      label: 'General Requests',
                    },
                    {
                      value: config?.rateLimit?.sandboxCreate?.limit || selectedOrganization?.sandboxCreateRateLimit,
                      label: 'Sandbox Creation',
                    },
                    {
                      value:
                        config?.rateLimit?.sandboxLifecycle?.limit || selectedOrganization?.sandboxLifecycleRateLimit,
                      label: 'Sandbox Lifecycle',
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
                    githubConnected: githubConnected,
                    businessEmailVerified: !!user?.profile?.business_email_verified,
                    phoneVerified: !!user?.profile?.phone_verified,
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
      </div>
    </div>
  )
}

function RateLimits({
  rateLimits,
  className,
}: {
  rateLimits: {
    value?: number | null
    label: string
  }[]
  className?: string
}) {
  const isEmpty = rateLimits.every(({ value }) => !value)
  if (isEmpty) {
    return null
  }

  return (
    <div className={cn('p-4 border-t border-border flex flex-col gap-4', className)}>
      <div className="flex flex-col gap-1">
        <div className="text-foreground text-sm font-medium">Rate Limits</div>
        <div className="text-muted-foreground text-sm">How many requests you can make per minute.</div>
      </div>
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
        {rateLimits.map(({ label, value }) => value && <RateLimitItem key={label} label={label} value={value} />)}
      </div>
    </div>
  )
}

function RateLimitItem({ label, value }: { label: string; value?: number | null }) {
  if (!value) {
    return null
  }

  return (
    <div className="flex flex-col">
      <div className="text-muted-foreground text-xs">{label}</div>
      <div className="text-foreground text-sm font-medium">{value?.toLocaleString()}</div>
    </div>
  )
}
