/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios, { AxiosInstance } from 'axios'
import { DaytonaError } from '@/api/errors'
import {
  AutomaticTopUp,
  OrganizationEmail,
  OrganizationTier,
  OrganizationUsage,
  OrganizationWallet,
  Tier,
} from './types'

export class BillingApiClient {
  private axiosInstance: AxiosInstance

  constructor(apiUrl: string, accessToken: string) {
    this.axiosInstance = axios.create({
      baseURL: apiUrl,
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })

    this.axiosInstance.interceptors.response.use(
      (response) => {
        return response
      },
      (error) => {
        const errorMessage = error.response?.data?.message || error.response?.data || error.message || String(error)

        throw DaytonaError.fromString(String(errorMessage))
      },
    )
  }

  public async getOrganizationUsage(organizationId: string): Promise<OrganizationUsage> {
    const response = await this.axiosInstance.get(`/organization/${organizationId}/usage`)
    return response.data
  }

  public async getPastOrganizationUsage(organizationId: string, periods?: number): Promise<OrganizationUsage[]> {
    const response = await this.axiosInstance.get(`/organization/${organizationId}/usage/past?periods=${periods || 12}`)
    return response.data
  }

  public async getOrganizationWallet(organizationId: string): Promise<OrganizationWallet> {
    const response = await this.axiosInstance.get(`/organization/${organizationId}/wallet`)
    return response.data
  }

  public async setAutomaticTopUp(organizationId: string, automaticTopUp?: AutomaticTopUp): Promise<void> {
    await this.axiosInstance.put(`/organization/${organizationId}/wallet/automatic-top-up`, automaticTopUp)
  }

  public async getOrganizationBillingPortalUrl(organizationId: string): Promise<string> {
    const response = await this.axiosInstance.get(`/organization/${organizationId}/portal-url`)
    return response.data
  }

  public async getOrganizationCheckoutUrl(organizationId: string): Promise<string> {
    const response = await this.axiosInstance.get(`/organization/${organizationId}/checkout-url`)
    return response.data
  }

  public async redeemCoupon(organizationId: string, couponCode: string): Promise<void> {
    await this.axiosInstance.post(`/organization/${organizationId}/redeem-coupon/${couponCode}`)
  }

  public async getOrganizationTier(organizationId: string): Promise<OrganizationTier> {
    const response = await this.axiosInstance.get(`/organization/${organizationId}/tier`)
    const orgTier: OrganizationTier = {
      tier: response.data.tier,
      largestSuccessfulPaymentDate: response.data.largestSuccessfulPaymentDate
        ? new Date(response.data.largestSuccessfulPaymentDate)
        : undefined,
      largestSuccessfulPaymentCents: response.data.largestSuccessfulPaymentCents,
      expiresAt: response.data.expiresAt ? new Date(response.data.expiresAt) : undefined,
      hasVerifiedBusinessEmail: response.data.hasVerifiedBusinessEmail,
    }

    return orgTier
  }

  public async upgradeTier(organizationId: string, tier: number): Promise<void> {
    await this.axiosInstance.post(`/organization/${organizationId}/tier/upgrade`, { tier })
  }

  public async downgradeTier(organizationId: string, tier: number): Promise<void> {
    await this.axiosInstance.post(`/organization/${organizationId}/tier/downgrade`, { tier })
  }

  public async listTiers(): Promise<Tier[]> {
    const response = await this.axiosInstance.get('/tier')
    return response.data
  }

  public async listOrganizationEmails(organizationId: string): Promise<OrganizationEmail[]> {
    const response = await this.axiosInstance.get(`/organization/${organizationId}/email`)
    return response.data.map((email: any) => ({
      ...email,
      verifiedAt: email.verifiedAt ? new Date(email.verifiedAt) : undefined,
    }))
  }

  public async addOrganizationEmail(organizationId: string, email: string): Promise<void> {
    await this.axiosInstance.post(`/organization/${organizationId}/email`, { email })
  }

  public async deleteOrganizationEmail(organizationId: string, email: string): Promise<void> {
    await this.axiosInstance.delete(`/organization/${organizationId}/email`, { data: { email } })
  }

  public async verifyOrganizationEmail(organizationId: string, email: string, token: string): Promise<void> {
    await this.axiosInstance.post(`/organization/${organizationId}/email/verify`, { email, token })
  }

  public async resendOrganizationEmailVerification(organizationId: string, email: string): Promise<void> {
    await this.axiosInstance.post(`/organization/${organizationId}/email/resend`, { email })
  }
}
