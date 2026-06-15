/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useMutation } from '@tanstack/react-query'
import { useApi } from '../useApi'

interface DownloadInvoiceVariables {
  organizationId: string
  invoiceId: string
}

export const useDownloadInvoiceMutation = () => {
  const { billingApi } = useApi()

  return useMutation<Blob, unknown, DownloadInvoiceVariables>({
    mutationFn: ({ organizationId, invoiceId }) => billingApi.downloadInvoice(organizationId, invoiceId),
  })
}
