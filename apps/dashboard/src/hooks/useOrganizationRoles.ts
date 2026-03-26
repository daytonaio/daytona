/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState, useEffect, useCallback } from 'react'
import { OrganizationRole } from '@daytonaio/api-client'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'

export const useOrganizationRoles = () => {
  const { organizationsApi } = useApi()

  const { selectedOrganization } = useSelectedOrganization()

  const [roles, setRoles] = useState<OrganizationRole[]>([])
  const [loadingRoles, setLoadingRoles] = useState(true)

  const fetchRoles = useCallback(
    async (showTableLoadingState = true) => {
      if (!selectedOrganization) {
        return
      }
      if (showTableLoadingState) {
        setLoadingRoles(true)
      }
      try {
        const response = await organizationsApi.listOrganizationRoles(selectedOrganization.id)
        setRoles(response.data)
      } catch (error) {
        handleApiError(error, 'Failed to fetch organization roles')
      } finally {
        setLoadingRoles(false)
      }
    },
    [organizationsApi, selectedOrganization],
  )

  useEffect(() => {
    fetchRoles()
  }, [fetchRoles])

  return {
    roles,
    loadingRoles,
    refreshRoles: fetchRoles,
  }
}
