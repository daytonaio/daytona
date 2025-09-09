/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TierAccordion, TierAccordionSkeleton } from '@/components/TierAccordion'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { UsageOverviewIndicator } from '@/components/UsageOverviewIndicator'
import { useDowngradeTierMutation } from '@/hooks/mutations/useDowngradeTierMutation'
import { useUpgradeTierMutation } from '@/hooks/mutations/useUpgradeTierMutation'
import { useOrganizationTierQuery } from '@/hooks/queries/useOrganizationTierQuery'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useOrganizationWalletQuery } from '@/hooks/queries/useOrganizationWalletQuery'
import { useTiersQuery } from '@/hooks/queries/useTiersQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { RefreshCcw } from 'lucide-react'
import React, { useCallback, useMemo } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'
import { UserProfileIdentity } from './LinkedAccounts'
import { useConfig } from '@/hooks/useConfig'


const Limits: React.FC = () => {
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const config = useConfig()

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
      refetchInterval: 10_000,
      refetchIntervalInBackground: true,
      staleTime: 0,
    },
  )

  const downgradeTier = useDowngradeTierMutation()
  const upgradeTier = useUpgradeTierMutation()

  const handleUpgradeTier = useCallback(
    async (tier: number) => {
      if (!selectedOrganization) {
        return
      }

      try {
        await upgradeTier.mutateAsync({ organizationId: selectedOrganization.id, tier })
        toast.success('Tier upgraded successfully')
      } catch (error) {
        handleApiError(error, 'Failed to upgrade organization tier')
      }
    },
    [upgradeTier, selectedOrganization],
  )

  const handleDowngradeTier = useCallback(
    async (tier: number) => {
      if (!selectedOrganization) {
        return
      }

      try {
        await downgradeTier.mutateAsync({ organizationId: selectedOrganization.id, tier })
        toast.success('Tier downgraded successfully')
      } catch (error) {
        handleApiError(error, 'Failed to downgrade organization tier')
      }
    },
    [downgradeTier, selectedOrganization],
  )

  const githubConnected = useMemo(() => {
    if (!user?.profile?.identities) {
      return false
    }
    return (user.profile.identities as UserProfileIdentity[]).some(
      (identity: UserProfileIdentity) => identity.provider === 'github',
    )
  }, [user])

  // const nextTierLabel = useMemo(() => {
  //   if (!organizationTier) {
  //     return null
  //   }

  //   const nextTier = tiers?.find((t) => t.tier === organizationTier.tier + 1)
  //   if (nextTier) {
  //     return `Tier ${nextTier.tier}`
  //   }

  //   return 'Custom Tier'
  // }, [organizationTier, tiers])

  // const hasHitLimit = useMemo(() => {
  //     if (!usageOverview) {
  //       return false
  //     }

  //     return pastUsage.some((u) => {
  //       return (
  //         u.peakCpuUsage >= usageOverview.totalCpuQuota ||
  //         u.peakMemUsage >= usageOverview.totalMemoryQuota ||
  //         u.peakDiskUsage >= usageOverview.totalDiskQuota
  //       )
  //     })
  //   }, [pastUsage, usageOverview])

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
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Limits</h1>
      </div>

      {/* todo: more granular error handling */}
      {isError ? (
        <Card className="my-4">
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
          <Card className="my-4">
            <CardHeader>
              <CardTitle className="flex justify-between gap-4 md:flex-row flex-col flex-wrap">
                <div className="flex items-center gap-2">
                  Current Usage{' '}
                  <Badge variant="outline" className="font-mono uppercase">
                    Tier {organizationTier?.tier}
                  </Badge>
                </div>
                {usageOverviewQuery.isLoading ? (
                  <Skeleton className="h-6 w-32" />
                ) : (
                  usageOverview && <UsageOverviewIndicator usage={usageOverview} isLive />
                )}
              </CardTitle>
            </CardHeader>

            {/* todo: implement past usage api */}
            {/* <div className="border border-border" />
          <CardContent>
            <LimitUsageChart
              title={
                <div className="py-6">
                  <CardTitle className="flex justify-between gap-4">Peak Usage</CardTitle>
                </div>
              }
              pastUsage={pastUsage}
              currentUsage={usageOverview}
            />
            {!!(hasHitLimit && nextTierLabel) && (
              <div className="text-sm text-muted-foreground mt-4">
                <TriangleAlertIcon className="inline-block h-4 align-middle text-yellow-600" />
                You hit your resource limit on one or more days this month. To ensure stable performance, upgrade to{' '}
                <span className="text-foreground">{nextTierLabel}</span>.
              </div>
            )}
          </CardContent> */}
          </Card>

          {config.billingApiUrl && (
            <Card className="my-4">
              <CardHeader>
                <CardTitle className="flex items-center mb-2">Upgrade your limits</CardTitle>
                <CardDescription>
                  Total vCPU, RAM, and Storage available across all active sandboxes. <br />
                  Usage depends on the compute needs of each sandbox.
                  <br />
                  <div className="text-sm text-muted-foreground/70 mt-2">
                    Note: for the top up requirements, make sure to top up in a single transaction.
                  </div>
                </CardDescription>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <TierAccordionSkeleton />
                ) : (
                  <TierAccordion
                    creditCardConnected={!!wallet?.creditCardConnected}
                    organizationTier={organizationTier}
                    emailVerified={!!user?.profile?.email_verified}
                    githubConnected={githubConnected}
                    tiers={tiers || []}
                    phoneVerified={!!user?.profile?.phone_verified}
                    tierLoading={organizationTierQuery.isLoading}
                    onUpgrade={handleUpgradeTier}
                    onDowngrade={handleDowngradeTier}
                  />
                )}
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  )
}

export default Limits