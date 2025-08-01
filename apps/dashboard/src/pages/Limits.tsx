/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Badge } from '@/components/ui/badge'
import { TierTable } from '@/components/TierTable'
import { OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import { useAuth } from 'react-oidc-context'
import { AlertTriangle } from 'lucide-react'
import { UsageOverview } from '@daytonaio/api-client'
import { handleApiError } from '@/lib/error-handling'
import QuotaLine from '@/components/QuotaLine'
import { Skeleton } from '@/components/ui/skeleton'
import { OrganizationTier, Tier } from '@/billing-api'
import { UserProfileIdentity } from './LinkedAccounts'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { toast } from 'sonner'
import { useConfig } from '@/hooks/useConfig'

const Limits: React.FC = () => {
  const { user } = useAuth()
  const { billingApi, organizationsApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const [organizationTier, setOrganizationTier] = useState<OrganizationTier | null>(null)
  const [tiers, setTiers] = useState<Tier[]>([])
  const [wallet, setWallet] = useState<OrganizationWallet | null>(null)
  const [usageOverview, setUsage] = useState<UsageOverview | null>(null)
  const [tierLoading, setTierLoading] = useState(false)
  const config = useConfig()

  const fetchOrganizationTier = useCallback(async () => {
    if (!config.billingApiUrl) {
      return
    }
    if (!selectedOrganization) {
      return
    }
    setTierLoading(true)
    try {
      const data = await billingApi.getOrganizationTier(selectedOrganization.id)
      setOrganizationTier(data)
    } catch (error) {
      handleApiError(error, 'Failed to fetch organization tier')
    } finally {
      setTierLoading(false)
    }
  }, [billingApi, selectedOrganization, config.billingApiUrl])

  const fetchTiers = useCallback(async () => {
    const data = await billingApi.listTiers()
    setTiers(data)
  }, [billingApi])

  const fetchOrganizationWallet = useCallback(async () => {
    if (!config.billingApiUrl) {
      return
    }
    if (!selectedOrganization) {
      return
    }
    try {
      const data = await billingApi.getOrganizationWallet(selectedOrganization.id)
      setWallet(data)
    } catch (error) {
      handleApiError(error, 'Failed to fetch organization wallet')
    }
  }, [billingApi, selectedOrganization, config.billingApiUrl])

  const fetchUsage = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    try {
      const response = await organizationsApi.getOrganizationUsageOverview(selectedOrganization.id)
      setUsage(response.data)
    } catch (error) {
      handleApiError(error, 'Failed to fetch usage data')
    }
  }, [organizationsApi, selectedOrganization])

  const upgradeTier = useCallback(
    async (tier: number) => {
      if (!selectedOrganization) {
        return
      }

      try {
        await billingApi.upgradeTier(selectedOrganization.id, tier)
        toast.success('Tier upgraded successfully')
        fetchOrganizationTier()
        fetchUsage()
      } catch (error) {
        handleApiError(error, 'Failed to upgrade organization tier')
      }
    },
    [billingApi, selectedOrganization, fetchOrganizationTier, fetchUsage],
  )

  const downgradeTier = useCallback(
    async (tier: number) => {
      if (!selectedOrganization) {
        return
      }

      try {
        await billingApi.downgradeTier(selectedOrganization.id, tier)
        toast.success('Tier downgraded successfully')
        fetchOrganizationTier()
        fetchUsage()
      } catch (error) {
        handleApiError(error, 'Failed to downgrade organization tier')
      }
    },
    [billingApi, selectedOrganization, fetchOrganizationTier, fetchUsage],
  )

  useEffect(() => {
    if (config.billingApiUrl) {
      // Fetch usage after tier because limits might have changed
      fetchOrganizationTier().finally(() => fetchUsage())
      fetchTiers()
    } else {
      fetchUsage()
    }
    const interval = setInterval(fetchUsage, 10000)
    return () => clearInterval(interval)
  }, [fetchOrganizationTier, fetchUsage, fetchTiers, config.billingApiUrl])

  useEffect(() => {
    fetchOrganizationWallet()
  }, [fetchOrganizationWallet])

  const getUsageDisplay = (current: number, total: number, unit = '') => {
    const percentage = (current / total) * 100
    const isHighUsage = percentage > 90

    return (
      <div className="flex items-center gap-1">
        <span className={isHighUsage ? 'text-red-500' : undefined}>
          {current} / {total} {unit}
        </span>
        {isHighUsage && <AlertTriangle className="w-4 h-4 text-red-500" />}
      </div>
    )
  }

  const githubConnected = useMemo(() => {
    if (!user?.profile?.identities) {
      return false
    }
    return (user.profile.identities as UserProfileIdentity[]).some(
      (identity: UserProfileIdentity) => identity.provider === 'github',
    )
  }, [user])

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Limits</h1>
      </div>

      <Card className="my-4">
        <CardHeader>
          <CardTitle className="flex items-center mb-2">
            Usage Limits{' '}
            {organizationTier && (
              <Badge variant="outline" className="ml-2 text-sm">
                Tier {organizationTier.tier}
              </Badge>
            )}
          </CardTitle>
          <CardDescription>
            Limits help us mitigate misuse and manage infrastructure resources. Ensuring fair and stable access to
            sandboxes and compute capacity across all users.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {!usageOverview && (
            <div className="flex items-center justify-center h-full">
              <Skeleton className="w-full h-full" />
            </div>
          )}
          {usageOverview && (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Resource</TableHead>
                  <TableHead>Usage</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow>
                  <TableCell>Compute</TableCell>
                  <TableCell>
                    <div className="max-w-80">
                      <div className="w-full flex justify-end">
                        {getUsageDisplay(usageOverview.currentCpuUsage, usageOverview.totalCpuQuota, 'vCPU')}
                      </div>
                      <QuotaLine current={usageOverview.currentCpuUsage} total={usageOverview.totalCpuQuota} />
                    </div>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Memory</TableCell>
                  <TableCell>
                    <div className="max-w-80">
                      <div className="w-full flex justify-end">
                        {getUsageDisplay(usageOverview.currentMemoryUsage, usageOverview.totalMemoryQuota, 'GiB')}
                      </div>
                      <QuotaLine current={usageOverview.currentMemoryUsage} total={usageOverview.totalMemoryQuota} />
                    </div>
                  </TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Storage</TableCell>
                  <TableCell>
                    <div className="max-w-80">
                      <div className="w-full flex justify-end">
                        {getUsageDisplay(usageOverview.currentDiskUsage, usageOverview.totalDiskQuota, 'GiB')}
                      </div>
                      <QuotaLine current={usageOverview.currentDiskUsage} total={usageOverview.totalDiskQuota} />
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {config.billingApiUrl && (
        <Card className="my-4">
          <CardHeader>
            <CardTitle className="flex items-center mb-2">Increasing your limits</CardTitle>
            {organizationTier && (
              <CardDescription>
                Your organization is currently in <b>Tier {organizationTier.tier}</b>. Your limits will automatically be
                increased once you move to the next tier based on the criteria outlined below.
              </CardDescription>
            )}
          </CardHeader>
          <CardContent>
            <TierTable
              creditCardConnected={!!wallet?.creditCardConnected}
              organizationTier={organizationTier}
              emailVerified={!!user?.profile?.email_verified}
              githubConnected={githubConnected}
              tiers={tiers}
              phoneVerified={!!user?.profile?.phone_verified}
              tierLoading={tierLoading}
              onUpgrade={upgradeTier}
              onDowngrade={downgradeTier}
            />
          </CardContent>
        </Card>
      )}
    </div>
  )
}

export default Limits
