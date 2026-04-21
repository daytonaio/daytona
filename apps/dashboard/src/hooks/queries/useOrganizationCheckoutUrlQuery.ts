/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useIsFetching, useQueryClient } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useBillingV2Enabled } from '../useBillingV2Enabled'
import { queryKeys } from './queryKeys'

export const useIsOrganizationCheckoutUrlFetching = (organizationId: string) => {
  const v2 = useBillingV2Enabled()
  return useIsFetching({ queryKey: queryKeys.billing.checkoutUrl(organizationId, v2) }) > 0
}

export const useFetchOrganizationCheckoutUrlQuery = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()
  const v2 = useBillingV2Enabled()

  return (organizationId: string) =>
    queryClient.fetchQuery({
      queryKey: queryKeys.billing.checkoutUrl(organizationId, v2),
      queryFn: () => billingApi.getOrganizationCheckoutUrl(organizationId, { v2 }),
    })
}
