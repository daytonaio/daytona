/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationEmail, OrganizationTier, OrganizationWallet } from '@/billing-api'
import { Tier } from '@/billing-api/types/tier'
import { DaytonaConfiguration } from '@daytonaio/api-client/src'
import { bypass, http, HttpResponse } from 'msw'

const BILLING_API_URL = 'http://localhost:3000/api/billing'
const API_URL = import.meta.env.VITE_API_URL

export const handlers = [
  http.get(`${API_URL}/config`, async () => {
    const originalConfig = await fetch(bypass(`${API_URL}/config`)).then((res) => res.json())

    return HttpResponse.json<Partial<DaytonaConfiguration>>({
      ...originalConfig,
      billingApiUrl: BILLING_API_URL,
    })
  }),
  http.get(`${BILLING_API_URL}/organization/:organizationId/portal-url`, async () => {
    return HttpResponse.json<string>(`${BILLING_API_URL}/portal`)
  }),
  http.get(`${BILLING_API_URL}/tier`, async () => {
    return HttpResponse.json<Tier[]>([
      {
        tier: 1,
        tierLimit: {
          concurrentCPU: 10,
          concurrentRAMGiB: 20,
          concurrentDiskGiB: 30,
        },
        minTopUpAmountCents: 0,
        topUpIntervalDays: 0,
      },
      {
        tier: 2,
        tierLimit: {
          concurrentCPU: 100,
          concurrentRAMGiB: 200,
          concurrentDiskGiB: 300,
        },
        minTopUpAmountCents: 2500,
        topUpIntervalDays: 0,
      },
      {
        tier: 3,
        tierLimit: {
          concurrentCPU: 250,
          concurrentRAMGiB: 500,
          concurrentDiskGiB: 2000,
        },
        minTopUpAmountCents: 50000,
        topUpIntervalDays: 0,
      },
      {
        tier: 4,
        tierLimit: {
          concurrentCPU: 500,
          concurrentRAMGiB: 1000,
          concurrentDiskGiB: 5000,
        },
        minTopUpAmountCents: 200000,
        topUpIntervalDays: 30,
      },
    ])
  }),
  http.get(`${BILLING_API_URL}/organization/:organizationId/wallet`, async () => {
    return HttpResponse.json<OrganizationWallet>({
      balanceCents: 1000,
      ongoingBalanceCents: 1000,
      name: 'Wallet',
      creditCardConnected: false,
      automaticTopUp: undefined,
    })
  }),
  http.get(`${BILLING_API_URL}/organization/:organizationId/tier`, async () => {
    return HttpResponse.json<OrganizationTier>({
      tier: 2,
      largestSuccessfulPaymentDate: new Date(),
      largestSuccessfulPaymentCents: 1000,
      expiresAt: new Date(),
      hasVerifiedBusinessEmail: true,
    })
  }),
  http.get(`${BILLING_API_URL}/organization/:organizationId/email`, async () => {
    return HttpResponse.json<OrganizationEmail[]>([
      {
        email: 'user@example.com',
        verified: true,
        owner: true,
        business: false,
        verifiedAt: new Date(),
      },
    ])
  }),
]
