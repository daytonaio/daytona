/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQueryClient } from '@tanstack/react-query'
import { useCallback } from 'react'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

// Returns a fetch function (not an eager query) so callers can trigger the
// URL fetch on demand — e.g. on button click — without prefetching a
// short-lived Stripe session on every page load.
export const useSetupCheckoutUrlQuery = (organizationId: string) => {
  const { billingApi } = useApi()
  const config = useConfig()
  const queryClient = useQueryClient()

  return useCallback(async (): Promise<string> => {
    // Mirror the billing availability guard used by the other billing queries:
    // without billingApiUrl the request would fall back to the dashboard origin.
    if (!config.billingApiUrl) {
      throw new Error('Billing is not available')
    }
    return queryClient.fetchQuery({
      queryKey: queryKeys.billing.setupCheckoutUrl(organizationId),
      queryFn: () => billingApi.getOrganizationSetupCheckoutUrl(organizationId),
      staleTime: 0,
      // A setup-checkout session is short-lived; retrying a single click could
      // spin up multiple Stripe sessions, so fail fast instead.
      retry: false,
    })
  }, [billingApi, config.billingApiUrl, organizationId, queryClient])
}
