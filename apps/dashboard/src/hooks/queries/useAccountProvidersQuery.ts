/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AccountProvider as AccountProviderApi } from '@daytonaio/api-client'
import { useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { queryKeys } from './queryKeys'

export const useAccountProvidersQuery = (
  options?: Omit<Parameters<typeof useQuery<AccountProviderApi[]>>[0], 'queryKey' | 'queryFn'>,
) => {
  const { userApi } = useApi()

  return useQuery<AccountProviderApi[]>({
    queryKey: queryKeys.user.accountProviders(),
    queryFn: async () => (await userApi.getAvailableAccountProviders()).data,
    ...options,
  })
}
