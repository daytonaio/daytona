/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IOrganizationsContext, OrganizationsContext } from '@/contexts/OrganizationsContext'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { useOrganizationsSuspenseQuery } from '@/hooks/queries/useOrganizationsQuery'
import { ReactNode, useCallback, useMemo } from 'react'

type Props = {
  children: ReactNode
}

export function OrganizationsProvider(props: Props) {
  const { data: organizations } = useOrganizationsSuspenseQuery()

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
      refreshOrganizations,
    }
  }, [organizations, refreshOrganizations])

  return <OrganizationsContext.Provider value={contextValue}>{props.children}</OrganizationsContext.Provider>
}
