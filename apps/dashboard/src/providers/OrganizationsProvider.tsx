/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ReactNode, useCallback, useMemo, useState } from 'react'
import { suspend } from 'suspend-react'
import { useApi } from '@/hooks/useApi'
import { OrganizationsContext, IOrganizationsContext } from '@/contexts/OrganizationsContext'
import { Organization } from '@daytonaio/api-client'
import { handleApiError } from '@/lib/error-handling'
import { LocalStorageKey } from '@/enums/LocalStorageKey'

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

  const refreshOrganizations = useCallback(
    async (selectedOrganizationId?: string) => {
      const orgs = await getOrganizations()
      setOrganizations(orgs)
      if (selectedOrganizationId) {
        localStorage.setItem(LocalStorageKey.SelectedOrganizationId, selectedOrganizationId)
      }
      // TODO: come back to this asap
      // After creating a new org, the selected org was updated unnecessarily so we reload the page just in case
      setTimeout(() => {
        window.location.reload()
      }, 500)
      return orgs
    },
    [getOrganizations],
  )

  const contextValue: IOrganizationsContext = useMemo(() => {
    return {
      organizations,
      setOrganizations,
      refreshOrganizations,
    }
  }, [organizations, refreshOrganizations])

  return <OrganizationsContext.Provider value={contextValue}>{props.children}</OrganizationsContext.Provider>
}
