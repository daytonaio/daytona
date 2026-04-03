/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PaginatedAuditLogs } from '@daytona/api-client'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApi } from '../useApi'
import { useSelectedOrganization } from '../useSelectedOrganization'
import { queryKeys } from './queryKeys'

export interface AuditLogsQueryParams {
  page: number
  pageSize: number
  from?: Date
  to?: Date
  cursor?: string
}

export function useAuditLogsQuery(
  params: AuditLogsQueryParams,
  options?: {
    enabled?: boolean
    refetchInterval?: number | false
  },
) {
  const { auditApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  return useQuery<PaginatedAuditLogs>({
    queryKey: queryKeys.audit.logs(selectedOrganization?.id ?? '', params),
    queryFn: async () => {
      if (!selectedOrganization) {
        throw new Error('No organization selected')
      }

      const response = await auditApi.getOrganizationAuditLogs(
        selectedOrganization.id,
        params.page,
        params.pageSize,
        params.from,
        params.to,
        params.cursor,
      )

      return response.data
    },
    enabled: Boolean(selectedOrganization && options?.enabled !== false),
    placeholderData: keepPreviousData,
    refetchInterval: options?.refetchInterval,
  })
}
