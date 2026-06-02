/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Charge } from '@daytona/billing-api-client'
import { useInfiniteQuery } from '@tanstack/react-query'
import { useEffect, useMemo } from 'react'
import { useApi } from '../useApi'
import { useConfig } from '../useConfig'
import { queryKeys } from './queryKeys'

export const useChargesQuery = ({ organizationId, enabled = true }: { organizationId: string; enabled?: boolean }) => {
  const { billingApi } = useApi()
  const config = useConfig()

  const query = useInfiniteQuery({
    queryKey: queryKeys.billing.charges(organizationId),
    queryFn: ({ pageParam }) => billingApi.listCharges(organizationId, { startingAfter: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => (lastPage.hasMore ? lastPage.nextCursor : undefined),
    enabled: Boolean(enabled && config.billingApiUrl && organizationId),
  })

  useEffect(() => {
    if (query.hasNextPage && !query.isFetchingNextPage) {
      query.fetchNextPage()
    }
  }, [query.hasNextPage, query.isFetchingNextPage, query.fetchNextPage])

  const charges = useMemo<Charge[]>(() => query.data?.pages.flatMap((page) => page.data ?? []) ?? [], [query.data])

  return {
    charges,
    isLoading: query.isLoading || query.hasNextPage || query.isFetchingNextPage,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,
  }
}
