/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ReactNode, useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import {
  UserOrganizationInvitationsContext,
  IUserOrganizationInvitationsContext,
} from '@/contexts/UserOrganizationInvitationsContext'

type Props = {
  children: ReactNode
}

export function UserOrganizationInvitationsProvider(props: Props) {
  const { organizationsApi } = useApi()

  const [count, setCount] = useState<number>(0)

  const getInvitationsCount = useCallback(async () => {
    try {
      const count = (await organizationsApi.getOrganizationInvitationsCountForAuthenticatedUser()).data
      setCount(count)
    } catch (e) {
      // silently fail
    }
  }, [organizationsApi])

  useEffect(() => {
    void getInvitationsCount()
  }, [getInvitationsCount])

  const contextValue: IUserOrganizationInvitationsContext = useMemo(() => {
    return {
      count,
      setCount,
    }
  }, [count])

  return (
    <UserOrganizationInvitationsContext.Provider value={contextValue}>
      {props.children}
    </UserOrganizationInvitationsContext.Provider>
  )
}
