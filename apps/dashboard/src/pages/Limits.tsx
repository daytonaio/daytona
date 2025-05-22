/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Badge } from '@/components/ui/badge'
import { TierTable } from '@/components/TierTable'

const Limits: React.FC = () => {
  const { billingApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const [organizationTier, setOrganizationTier] = useState<number | null>(null)

  const fetchOrganizationTier = useCallback(async () => {
    if (!import.meta.env.VITE_BILLING_API_URL) {
      return
    }
    if (!selectedOrganization) {
      return
    }
    try {
      const data = await billingApi.getOrganizationTier(selectedOrganization.id)
      setOrganizationTier(data)
    } catch (error) {
      console.error('Failed to fetch organization tier:', error)
    }
  }, [billingApi, selectedOrganization])

  useEffect(() => {
    fetchOrganizationTier()
  }, [fetchOrganizationTier])

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold">Limits</h1>

      {import.meta.env.VITE_BILLING_API_URL && (
        <Card className="my-4">
          <CardHeader>
            <CardTitle className="flex items-center">
              Tiers
              {organizationTier && (
                <Badge variant="outline" className="ml-2 text-sm">
                  Current Tier: {organizationTier}
                </Badge>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className="p-6">
            <TierTable />
          </CardContent>
        </Card>
      )}
    </div>
  )
}

export default Limits
