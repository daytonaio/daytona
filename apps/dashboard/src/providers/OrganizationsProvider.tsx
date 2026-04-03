/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IOrganizationsContext, OrganizationsContext } from '@/contexts/OrganizationsContext'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { useApi } from '@/hooks/useApi'
import { handleApiError } from '@/lib/error-handling'
import { Organization } from '@daytona/api-client'
import { ReactNode, useCallback, useMemo, useState } from 'react'
import { suspend } from 'suspend-react'

type Props = {
  children: ReactNode
}

export function OrganizationsProvider(props: Props) {
  const { organizationsApi } = useApi()

  const getOrganizations = useCallback(async () => {
    try {
      return (await organizationsApi.listOrganizations()).data
    } catch (error) {
      handleApiError(error, 'Failed to fetch your organizations')
      throw error
    }
  }, [organizationsApi])

  const [organizations, setOrganizations] = useState<Organization[]>(
    suspend(getOrganizations, [organizationsApi, 'organizations']),
  )

  // TODO: come back to this asap
  const refreshOrganizations_todo = useCallback(
    async (selectedOrganizationId?: string) => {
      const orgs = await getOrganizations()
      setOrganizations(orgs)
      if (selectedOrganizationId) {
        localStorage.setItem(LocalStorageKey.SelectedOrganizationId, selectedOrganizationId)
      }
    },
    [getOrganizations],
  )

  // After creating a new org, the selected org was updated unnecessarily so we reload the page just in case
  const refreshOrganizations = useCallback(async (selectedOrganizationId?: string) => {
    return new Promise<void>(() => {
      if (selectedOrganizationId) {
        localStorage.setItem(LocalStorageKey.SelectedOrganizationId, selectedOrganizationId)
      }
      setTimeout(() => {
        window.location.reload()
      }, 500)
    })
  }, [])

  const contextValue: IOrganizationsContext = useMemo(() => {
    return {
      organizations,
      setOrganizations,
      refreshOrganizations,
    }
  }, [organizations, refreshOrganizations])

  return <OrganizationsContext.Provider value={contextValue}>{props.children}</OrganizationsContext.Provider>
}
