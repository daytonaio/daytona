/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  AutomaticTopUp,
  BillingInfo,
  BillingInfoApi,
  ChargeList,
  CheckoutUrlApi,
  Configuration,
  InvoicesApi,
  OrganizationApi,
  OrganizationEmail,
  OrganizationTier,
  OrganizationUsage,
  OrganizationWallet,
  PaginatedTInvoice,
  PaymentMethod,
  PaymentUrl,
  PortalUrlApi,
  Tier,
  TierApi,
  UsageApi,
  WalletApi,
} from '@daytona/billing-api-client'
import { AxiosInstance } from 'axios'

// Passed per-call by hooks that read the BILLING_PROVIDER_V2 feature flag.
// Each method picks the v1 or v2 backend based on this flag. Methods without
// a v2 counterpart (tiers, emails) ignore it.
export interface BillingVersionOptions {
  v2?: boolean
}

export class BillingApiClient {
  private walletApi: WalletApi
  private usageApi: UsageApi
  private tierApi: TierApi
  private invoicesApi: InvoicesApi
  private organizationApi: OrganizationApi
  private checkoutUrlApi: CheckoutUrlApi
  private portalUrlApi: PortalUrlApi
  private billingInfoApi: BillingInfoApi

  constructor(configuration: Configuration, axiosInstance: AxiosInstance) {
    this.walletApi = new WalletApi(configuration, undefined, axiosInstance)
    this.usageApi = new UsageApi(configuration, undefined, axiosInstance)
    this.tierApi = new TierApi(configuration, undefined, axiosInstance)
    this.invoicesApi = new InvoicesApi(configuration, undefined, axiosInstance)
    this.organizationApi = new OrganizationApi(configuration, undefined, axiosInstance)
    this.checkoutUrlApi = new CheckoutUrlApi(configuration, undefined, axiosInstance)
    this.portalUrlApi = new PortalUrlApi(configuration, undefined, axiosInstance)
    this.billingInfoApi = new BillingInfoApi(configuration, undefined, axiosInstance)
  }

  public async getOrganizationUsage(
    organizationId: string,
    { v2 }: BillingVersionOptions = {},
  ): Promise<OrganizationUsage> {
    const response = v2
      ? await this.usageApi.getV2CurrentUsage(organizationId)
      : await this.usageApi.getCurrentUsage(organizationId)
    return response.data
  }

  public async getPastOrganizationUsage(
    organizationId: string,
    periods?: number,
    { v2 }: BillingVersionOptions = {},
  ): Promise<OrganizationUsage[]> {
    const response = v2
      ? await this.usageApi.getV2PastUsage(organizationId, periods ?? 12)
      : await this.usageApi.getPastUsage(organizationId, periods ?? 12)
    // v1 returns a single OrganizationUsage; v2 returns an array. Normalize to array.
    const data = response.data as unknown as OrganizationUsage | OrganizationUsage[]
    return Array.isArray(data) ? data : [data]
  }

  public async getOrganizationWallet(
    organizationId: string,
    { v2 }: BillingVersionOptions = {},
  ): Promise<OrganizationWallet> {
    const response = v2
      ? await this.walletApi.getV2Wallet(organizationId)
      : await this.walletApi.getWallet(organizationId)
    return response.data
  }

  public async setAutomaticTopUp(
    organizationId: string,
    automaticTopUp?: AutomaticTopUp,
    { v2 }: BillingVersionOptions = {},
  ): Promise<void> {
    if (v2) {
      await this.walletApi.setV2AutomaticTopUp(organizationId, automaticTopUp)
    } else {
      await this.walletApi.setAutomaticTopUp(organizationId, automaticTopUp)
    }
  }

  public async getOrganizationBillingPortalUrl(
    organizationId: string,
    { v2 }: BillingVersionOptions = {},
  ): Promise<string> {
    const response = v2
      ? await this.portalUrlApi.getV2PortalURL(organizationId)
      : await this.portalUrlApi.getPortalUrl(organizationId)
    return response.data
  }

  public async getOrganizationCheckoutUrl(organizationId: string, { v2 }: BillingVersionOptions = {}): Promise<string> {
    const response = v2
      ? await this.checkoutUrlApi.getV2CheckoutURL(organizationId)
      : await this.checkoutUrlApi.getCheckoutUrl(organizationId)
    return response.data
  }

  public async redeemCoupon(
    organizationId: string,
    couponCode: string,
    { v2 }: BillingVersionOptions = {},
  ): Promise<string> {
    if (v2) {
      await this.organizationApi.redeemV2Coupon(couponCode, organizationId)
    } else {
      await this.organizationApi.redeemCoupon(couponCode, organizationId)
    }
    return 'Coupon redeemed successfully'
  }

  public async listInvoices(
    organizationId: string,
    page?: number,
    perPage?: number,
    { v2 }: BillingVersionOptions = {},
  ): Promise<PaginatedTInvoice> {
    const response = v2
      ? await this.invoicesApi.listV2Invoices(organizationId, page, perPage)
      : await this.invoicesApi.listInvoices(organizationId, page, perPage)
    return response.data
  }

  public async createInvoicePaymentUrl(
    organizationId: string,
    invoiceId: string,
    { v2 }: BillingVersionOptions = {},
  ): Promise<PaymentUrl> {
    const response = v2
      ? await this.invoicesApi.createV2PaymentURL(organizationId, invoiceId)
      : await this.invoicesApi.createPaymentUrl(organizationId, invoiceId)
    return response.data
  }

  public async voidInvoice(
    organizationId: string,
    invoiceId: string,
    { v2 }: BillingVersionOptions = {},
  ): Promise<void> {
    if (v2) {
      throw new Error('Voiding invoices is not supported in billing provider v2')
    } else {
      await this.invoicesApi.voidInvoice(organizationId, invoiceId)
    }
  }

  public async topUpWallet(
    organizationId: string,
    amountCents: number,
    { v2 }: BillingVersionOptions = {},
  ): Promise<PaymentUrl> {
    const response = v2
      ? await this.walletApi.topUpV2Wallet(organizationId, { amountCents })
      : await this.walletApi.topUpWallet(organizationId, { amountCents })
    return response.data
  }

  // Tier + email endpoints have no v2 counterpart; they always call v1.

  public async getOrganizationTier(organizationId: string): Promise<OrganizationTier> {
    const response = await this.organizationApi.getTier(organizationId)
    return response.data
  }

  public async upgradeTier(organizationId: string, tier: number): Promise<void> {
    await this.organizationApi.upgradeTier(organizationId, { tier })
  }

  public async downgradeTier(organizationId: string, tier: number): Promise<void> {
    await this.organizationApi.downgradeTier(organizationId, { tier })
  }

  public async listTiers(): Promise<Tier[]> {
    const response = await this.tierApi.listTiers()
    return response.data
  }

  public async listOrganizationEmails(organizationId: string): Promise<OrganizationEmail[]> {
    const response = await this.organizationApi.listOrganizationEmails(organizationId)
    return response.data
  }

  public async addOrganizationEmail(organizationId: string, email: string): Promise<void> {
    await this.organizationApi.addOrganizationEmail(organizationId, { email })
  }

  public async deleteOrganizationEmail(organizationId: string, email: string): Promise<void> {
    await this.organizationApi.deleteOrganizationEmail(organizationId, { email })
  }

  public async verifyOrganizationEmail(organizationId: string, email: string, token: string): Promise<void> {
    await this.organizationApi.verifyEmail(organizationId, { email, token })
  }

  public async resendOrganizationEmailVerification(organizationId: string, email: string): Promise<void> {
    await this.organizationApi.resendVerificationEmail(organizationId, { email })
  }

  // v2-only: billing info, payment methods, charges. These have no v1 counterpart
  // and are only invoked when the BILLING_PROVIDER_V2 flag is on.

  public async getBillingInfo(organizationId: string): Promise<BillingInfo> {
    const response = await this.billingInfoApi.getV2BillingInfo(organizationId)
    return response.data
  }

  public async listPaymentMethods(organizationId: string): Promise<PaymentMethod[]> {
    const response = await this.billingInfoApi.listV2PaymentMethods(organizationId)
    return response.data
  }

  public async listCharges(
    organizationId: string,
    { limit, startingAfter }: { limit?: number; startingAfter?: string } = {},
  ): Promise<ChargeList> {
    const response = await this.billingInfoApi.listV2Charges(organizationId, limit, startingAfter)
    return response.data
  }
}
