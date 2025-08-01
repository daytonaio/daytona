/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createContext } from 'react'
import { OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import { OrganizationTier } from '@/billing-api/types/OrganizationTier'
import { Tier } from '@/billing-api/types/tier'
import { OrganizationEmail } from '@/billing-api'

export interface IBillingContext {
  // Wallet data
  wallet: OrganizationWallet | null
  walletLoading: boolean
  refreshWallet: () => Promise<void>

  // Tier data
  organizationTier: OrganizationTier | null
  tierLoading: boolean
  tiers: Tier[]
  tiersLoading: boolean
  refreshTier: () => Promise<void>
  refreshTiers: () => Promise<void>

  // Organization emails
  organizationEmails: OrganizationEmail[]
  organizationEmailsLoading: boolean
  refreshOrganizationEmails: () => Promise<void>

  // Billing portal
  billingPortalUrl: string | null
  billingPortalUrlLoading: boolean
  refreshBillingPortalUrl: () => Promise<void>
}

export const BillingContext = createContext<IBillingContext | null>(null)
