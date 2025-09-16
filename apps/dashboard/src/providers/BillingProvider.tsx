'use client'

/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type React from 'react'
import { useEffect, useState, useCallback, type ReactNode } from 'react'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import type { OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import type { OrganizationTier } from '@/billing-api/types/OrganizationTier'
import type { Tier } from '@/billing-api/types/tier'
import type { OrganizationEmail } from '@/billing-api'
import { BillingContext, type IBillingContext } from '@/contexts/BillingContext'
import { handleApiError } from '@/lib/error-handling'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'

interface BillingProviderProps {
  children: ReactNode
}

const MOCK_WALLET: OrganizationWallet = {
  balanceCents: 15075, // $150.75 in cents
  ongoingBalanceCents: 2500, // $25.00 in ongoing charges
  name: 'Development Wallet',
  creditCardConnected: true,
  automaticTopUp: {
    thresholdAmount: 1000, // $10.00 threshold
    targetAmount: 5000, // $50.00 target
  },
}

const MOCK_ORGANIZATION_TIER: OrganizationTier = {
  tier: 2, // Pro tier level
  largestSuccessfulPaymentDate: new Date('2024-01-15'),
  largestSuccessfulPaymentCents: 2900, // $29.00 in cents
  expiresAt: new Date('2024-12-31'),
  hasVerifiedBusinessEmail: true,
}

const MOCK_TIERS: Tier[] = [
  {
    tier: 0, // Free tier
    tierLimit: {
      concurrentCPU: 2,
      concurrentRAMGiB: 4,
      concurrentDiskGiB: 10,
    },
    minTopUpAmountCents: 1000, // $10.00 minimum top-up
    topUpIntervalDays: 30,
  },
  {
    tier: 1, // Pro tier
    tierLimit: {
      concurrentCPU: 8,
      concurrentRAMGiB: 16,
      concurrentDiskGiB: 100,
    },
    minTopUpAmountCents: 2000, // $20.00 minimum top-up
    topUpIntervalDays: 30,
  },
  {
    tier: 2, // Enterprise tier
    tierLimit: {
      concurrentCPU: 32,
      concurrentRAMGiB: 64,
      concurrentDiskGiB: 500,
    },
    minTopUpAmountCents: 5000, // $50.00 minimum top-up
    topUpIntervalDays: 30,
  },
]

const MOCK_ORGANIZATION_EMAILS: OrganizationEmail[] = [
  {
    email: 'billing@example.com',
    verified: true,
    owner: true,
    business: true,
    verifiedAt: new Date('2024-01-01'),
  },
  {
    email: 'admin@example.com',
    verified: true,
    owner: false,
    business: false,
    verifiedAt: new Date('2024-01-15'),
  },
]

const MOCK_BILLING_PORTAL_URL = 'https://billing.stripe.com/p/session/test_mock_session_id'

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

      // Return mock data for local development
      if (import.meta.env.VITE_BILLING_API_URL === 'http://localhost:6100') {
        console.log('[MOCK] Returning mock wallet data')
        return MOCK_WALLET
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

      // Return mock data for local development
      if (import.meta.env.VITE_BILLING_API_URL === 'http://localhost:6100') {
        console.log('[MOCK] Returning mock organization tier data')
        return MOCK_ORGANIZATION_TIER
      }

      return await billingApi.getOrganizationTier(selectedOrganizationId)
    },
    [billingApi, isOwner],
  )

  const getTiers = useCallback(async () => {
    if (!import.meta.env.VITE_BILLING_API_URL) {
      return []
    }

    // Return mock data for local development
    if (import.meta.env.VITE_BILLING_API_URL === 'http://localhost:6100') {
      console.log('[MOCK] Returning mock tiers data')
      return MOCK_TIERS
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

      // Return mock data for local development
      if (import.meta.env.VITE_BILLING_API_URL === 'http://localhost:6100') {
        console.log('[MOCK] Returning mock organization emails data')
        return MOCK_ORGANIZATION_EMAILS
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

      // Return mock data for local development
      if (import.meta.env.VITE_BILLING_API_URL === 'http://localhost:6100') {
        console.log('[MOCK] Returning mock billing portal URL')
        return MOCK_BILLING_PORTAL_URL
      }

      return await billingApi.getOrganizationBillingPortalUrl(selectedOrganizationId)
    },
    [billingApi, isOwner],
  )

  // Wallet state
  const [wallet, setWallet] = useState<OrganizationWallet | null>(null)
  const [walletLoading, setWalletLoading] = useState(true)

  // Tier state
  const [organizationTier, setOrganizationTier] = useState<OrganizationTier | null>(null)
  const [tierLoading, setTierLoading] = useState(false)
  const [tiers, setTiers] = useState<Tier[]>([])
  const [tiersLoading, setTiersLoading] = useState(false)

  // Organization emails state
  const [organizationEmails, setOrganizationEmails] = useState<OrganizationEmail[]>([])
  const [organizationEmailsLoading, setOrganizationEmailsLoading] = useState(true)

  // Billing portal state
  const [billingPortalUrl, setBillingPortalUrl] = useState<string | null>(null)
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
