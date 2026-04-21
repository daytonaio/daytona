/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Daytona } from '@daytona/sdk'
import { useMemo } from 'react'
import { useAuth } from 'react-oidc-context'

import { useConfig } from '@/hooks/useConfig'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'

import { useSandboxInstanceQuery } from './queries'

export function useSandboxInstance(sandboxId: string) {
  const { user } = useAuth()
  const { apiUrl } = useConfig()
  const { selectedOrganization } = useSelectedOrganization()

  const client = useMemo(() => {
    if (!user?.access_token || !selectedOrganization?.id) {
      return null
    }

    return new Daytona({
      jwtToken: user.access_token,
      apiUrl,
      organizationId: selectedOrganization.id,
    })
  }, [apiUrl, selectedOrganization?.id, user?.access_token])

  return useSandboxInstanceQuery({
    client,
    sandboxId,
  })
}
