/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { LiveIndicator } from '@/components/LiveIndicator'
import { TierAccordion, TierAccordionSkeleton } from '@/components/TierAccordion'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { UsageOverview, UsageOverviewSkeleton } from '@/components/UsageOverview'
import { useDowngradeTierMutation } from '@/hooks/mutations/useDowngradeTierMutation'
import { useUpgradeTierMutation } from '@/hooks/mutations/useUpgradeTierMutation'
import { useOrganizationTierQuery } from '@/hooks/queries/useOrganizationTierQuery'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useOrganizationWalletQuery } from '@/hooks/queries/useOrganizationWalletQuery'
import { useTiersQuery } from '@/hooks/queries/useTiersQuery'
import { useConfig } from '@/hooks/useConfig'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { keepPreviousData } from '@tanstack/react-query'
import { RefreshCcw } from 'lucide-react'
import React, { useCallback, useMemo } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'
import { UserProfileIdentity } from './LinkedAccounts'

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
      placeholderData: keepPreviousData,
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
              <CardTitle className="flex justify-between gap-x-4 gap-y-2 flex-row flex-wrap">
                <div className="flex items-center gap-2">
                  Current Usage{' '}
                  {organizationTier && (
                    <Badge variant="outline" className="font-mono uppercase">
                      Tier {organizationTier.tier}
                    </Badge>
                  )}
                </div>
                {usageOverview && (
                  <LiveIndicator
                    isUpdating={usageOverviewQuery.isFetching}
                    intervalMs={10_000}
                    lastUpdatedAt={usageOverviewQuery.dataUpdatedAt || 0}
                  />
                )}
              </CardTitle>
              <CardDescription>
                Limits help us mitigate misuse and manage infrastructure resources. <br /> Ensuring fair and stable
                access to sandboxes and compute capacity across all users.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {usageOverviewQuery.isLoading ? (
                <UsageOverviewSkeleton />
              ) : (
                usageOverview && <UsageOverview usageOverview={usageOverview} />
              )}
            </CardContent>
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
              <CardContent className="p-0 mt-4">
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
