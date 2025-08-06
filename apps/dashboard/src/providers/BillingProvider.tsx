/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState, useCallback, ReactNode } from 'react'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import { OrganizationTier } from '@/billing-api/types/OrganizationTier'
import { Tier } from '@/billing-api/types/tier'
import { OrganizationEmail } from '@/billing-api'
import { BillingContext, IBillingContext } from '@/contexts/BillingContext'
import { handleApiError } from '@/lib/error-handling'
import { suspend } from 'suspend-react'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'

interface BillingProviderProps {
  children: ReactNode
}

export const BillingProvider: React.FC<BillingProviderProps> = ({ children }) => {
  const { billingApi } = useApi()
  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()

  // Helper function to check if user has owner role
  const isOwner = useCallback(() => {
    return authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER
  }, [authenticatedUserOrganizationMember])

  const getWallet = useCallback(
    async (selectedOrganizationId?: string) => {
      if (!import.meta.env.VITE_BILLING_API_URL || !selectedOrganizationId || !isOwner()) {
        return null
      }
      try {
        return await billingApi.getOrganizationWallet(selectedOrganizationId)
      } catch (error) {
        handleApiError(error, 'Failed to fetch wallet data')
        throw error
      }
    },
    [billingApi, isOwner],
  )

  const getOrganizationTier = useCallback(
    async (selectedOrganizationId?: string) => {
      if (!import.meta.env.VITE_BILLING_API_URL || !selectedOrganizationId || !isOwner()) {
        return null
      }
      return await billingApi.getOrganizationTier(selectedOrganizationId)
    },
    [billingApi, isOwner],
  )

  const getTiers = useCallback(async () => {
    if (!import.meta.env.VITE_BILLING_API_URL) {
      return []
    }
    try {
      return await billingApi.listTiers()
    } catch (error) {
      handleApiError(error, 'Failed to fetch tiers')
      throw error
    }
  }, [billingApi])

  const getOrganizationEmails = useCallback(
    async (selectedOrganizationId?: string) => {
      if (!import.meta.env.VITE_BILLING_API_URL || !selectedOrganizationId || !isOwner()) {
        return []
      }
      return await billingApi.listOrganizationEmails(selectedOrganizationId)
    },
    [billingApi, isOwner],
  )

  const getBillingPortalUrl = useCallback(
    async (selectedOrganizationId?: string) => {
      if (!import.meta.env.VITE_BILLING_API_URL || !selectedOrganizationId || !isOwner()) {
        return null
      }
      return await billingApi.getOrganizationBillingPortalUrl(selectedOrganizationId)
    },
    [billingApi, isOwner],
  )

  // Wallet state
  const [wallet, setWallet] = useState<OrganizationWallet | null>(
    suspend(() => getWallet(selectedOrganization?.id), ['wallet']),
  )
  const [walletLoading, setWalletLoading] = useState(true)

  // Tier state
  const [organizationTier, setOrganizationTier] = useState<OrganizationTier | null>(() =>
    suspend(() => getOrganizationTier(selectedOrganization?.id), ['organizationTier']),
  )
  const [tierLoading, setTierLoading] = useState(false)

  const [tiers, setTiers] = useState<Tier[]>(() => suspend(() => getTiers(), ['tiers']))
  const [tiersLoading, setTiersLoading] = useState(false)

  // Organization emails state
  const [organizationEmails, setOrganizationEmails] = useState<OrganizationEmail[]>(() =>
    suspend(() => getOrganizationEmails(selectedOrganization?.id), ['organizationEmails']),
  )
  const [organizationEmailsLoading, setOrganizationEmailsLoading] = useState(true)

  // Billing portal state
  const [billingPortalUrl, setBillingPortalUrl] = useState<string | null>(() =>
    suspend(() => getBillingPortalUrl(selectedOrganization?.id), ['billingPortalUrl']),
  )
  const [billingPortalUrlLoading, setBillingPortalUrlLoading] = useState(true)

  // Fetch wallet data
  const refreshWallet = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setWalletLoading(true)
    try {
      const data = await getWallet(selectedOrganization.id)
      setWallet(data)
    } catch {
      setWallet(null)
    } finally {
      setWalletLoading(false)
    }
  }, [getWallet, selectedOrganization])

  // Fetch tier data
  const refreshTier = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setTierLoading(true)
    try {
      const data = await getOrganizationTier(selectedOrganization.id)
      setOrganizationTier(data)
    } catch {
      setOrganizationTier(null)
    } finally {
      setTierLoading(false)
    }
  }, [getOrganizationTier, selectedOrganization])

  // Fetch available tiers
  const refreshTiers = useCallback(async () => {
    setTiersLoading(true)
    try {
      const data = await getTiers()
      setTiers(data)
    } catch {
      setTiers([])
    } finally {
      setTiersLoading(false)
    }
  }, [getTiers])

  // Fetch organization emails
  const refreshOrganizationEmails = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setOrganizationEmailsLoading(true)
    try {
      const data = await getOrganizationEmails(selectedOrganization.id)
      setOrganizationEmails(data)
    } catch {
      setOrganizationEmails([])
    } finally {
      setOrganizationEmailsLoading(false)
    }
  }, [getOrganizationEmails, selectedOrganization])

  // Fetch billing portal URL
  const refreshBillingPortalUrl = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setBillingPortalUrlLoading(true)
    try {
      const data = await getBillingPortalUrl(selectedOrganization.id)
      setBillingPortalUrl(data)
    } catch (error) {
      handleApiError(error, 'Failed to fetch billing portal url')
      setBillingPortalUrl(null)
    } finally {
      setBillingPortalUrlLoading(false)
    }
  }, [getBillingPortalUrl, selectedOrganization])

  // Initialize data when organization changes
  useEffect(() => {
    if (!import.meta.env.VITE_BILLING_API_URL) {
      return
    }

    if (selectedOrganization && isOwner()) {
      refreshWallet()
      refreshTier()
      refreshTiers()
      refreshOrganizationEmails()
      refreshBillingPortalUrl()
    } else {
      // Reset state when no organization is selected
      setWallet(null)
      setOrganizationTier(null)
      setTiers([])
      setOrganizationEmails([])
      setBillingPortalUrl(null)
    }
  }, [
    selectedOrganization,
    isOwner,
    refreshWallet,
    refreshTier,
    refreshTiers,
    refreshOrganizationEmails,
    refreshBillingPortalUrl,
  ])

  const contextValue: IBillingContext = {
    // Wallet data
    wallet,
    walletLoading,
    refreshWallet,

    // Tier data
    organizationTier,
    tierLoading,
    tiers,
    tiersLoading,
    refreshTier,
    refreshTiers,

    // Organization emails
    organizationEmails,
    organizationEmailsLoading,
    refreshOrganizationEmails,

    // Billing portal
    billingPortalUrl,
    billingPortalUrlLoading,
    refreshBillingPortalUrl,
  }

  return <BillingContext.Provider value={contextValue}>{children}</BillingContext.Provider>
}
