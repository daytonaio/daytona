/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export interface InvoiceErrorDetail {
  details?: Record<string, string>
  errorCode: string
}

export type InvoicePaymentStatus = 'pending' | 'succeeded' | 'failed'
export type InvoiceStatus = 'draft' | 'finalized' | 'failed' | 'voided' | 'pending'
export type InvoiceType = 'subscription' | 'add_on' | 'one_off'

export interface Invoice {
  currency: string
  errorDetails?: InvoiceErrorDetail[]
  fileUrl?: string
  id: string
  issuingDate: string
  number: string
  paymentDueDate: string
  paymentOverdue: boolean
  paymentStatus: InvoicePaymentStatus
  sequentialId: number
  status: InvoiceStatus
  totalAmountCents: number
  totalDueAmountCents: number
  type: InvoiceType
}

export interface PaginatedInvoices {
  items: Invoice[]
  totalItems: number
  totalPages: number
}

export interface PaymentUrl {
  url: string
}
