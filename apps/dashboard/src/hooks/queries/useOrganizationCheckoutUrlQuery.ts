/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useIsFetching, useQueryClient } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

export const useIsOrganizationCheckoutUrlFetching = (organizationId: string) =>
  useIsFetching({ queryKey: queryKeys.billing.checkoutUrl(organizationId) }) > 0

export const useFetchOrganizationCheckoutUrlQuery = () => {
  const { billingApi } = useApi()
  const queryClient = useQueryClient()

  return (organizationId: string) =>
    queryClient.fetchQuery({
      queryKey: queryKeys.billing.checkoutUrl(organizationId),
      queryFn: () => billingApi.getOrganizationCheckoutUrl(organizationId),
    })
}
