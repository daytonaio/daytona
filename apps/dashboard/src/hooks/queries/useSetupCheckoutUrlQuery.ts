/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQueryClient } from '@tanstack/react-query'
import { useCallback } from 'react'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

// Returns a fetch function (not an eager query) so callers can trigger the
// URL fetch on demand — e.g. on button click — without prefetching a
// short-lived Stripe session on every page load.
export const useSetupCheckoutUrlQuery = (organizationId: string) => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return useCallback(async (): Promise<string> => {
    return queryClient.fetchQuery({
      queryKey: queryKeys.billing.setupCheckoutUrl(organizationId),
      queryFn: () => billingApi.getOrganizationSetupCheckoutUrl(organizationId),
      staleTime: 0,
    })
  }, [billingApi, organizationId, queryClient])
}
