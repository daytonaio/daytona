/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQueries } from '@tanstack/react-query'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { RunnerClass } from '@daytonaio/api-client'
import { useMemo } from 'react'

interface UseSnapshotRunnerClassesResult {
  runnerClassMap: Record<string, RunnerClass>
  isLoading: boolean
}

export function useSnapshotRunnerClasses(snapshotNames: string[]): UseSnapshotRunnerClassesResult {
  const { snapshotApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  // Deduplicate snapshot names
  const uniqueSnapshotNames = useMemo(() => [...new Set(snapshotNames.filter(Boolean))], [snapshotNames])

  const queries = useQueries({
    queries: uniqueSnapshotNames.map((name) => ({
      queryKey: ['snapshot-runner-class', selectedOrganization?.id, name],
      queryFn: async () => {
        if (!selectedOrganization) {
          throw new Error('No organization selected')
        }
        const response = await snapshotApi.getSnapshotRunnerClass(name, selectedOrganization.id)
        return { name, runnerClass: response.data.runnerClass }
      },
      enabled: !!selectedOrganization && !!name,
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10, // 10 minutes
    })),
  })

  const runnerClassMap = useMemo(() => {
    const map: Record<string, RunnerClass> = {}
    queries.forEach((query) => {
      if (query.data) {
        map[query.data.name] = query.data.runnerClass
      }
    })
    return map
  }, [queries])

  const isLoading = queries.some((query) => query.isLoading)

  return { runnerClassMap, isLoading }
}
