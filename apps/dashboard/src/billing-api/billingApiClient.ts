/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios, { AxiosError, AxiosInstance } from 'axios'
import { OrganizationUsage } from './types/OrganizationUsage'
import { AutomaticTopUp, OrganizationWallet } from './types/OrganizationWallet'
import { DaytonaError } from '@/api/errors'

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
}
