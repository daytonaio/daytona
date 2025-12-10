/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RoutePath } from '@/enums/RoutePath'
import { queryKeys } from '@/hooks/queries/queryKeys'
import { DaytonaConfiguration } from '@daytonaio/api-client'
import { useSuspenseQuery } from '@tanstack/react-query'
import { InMemoryWebStorage, WebStorageStateStore } from 'oidc-client-ts'
import { ReactNode, useMemo } from 'react'
import { AuthProvider, AuthProviderProps } from 'react-oidc-context'
import { ConfigContext } from '../contexts/ConfigContext'

const apiUrl = (import.meta.env.VITE_BASE_API_URL ?? window.location.origin) + '/api'

type Props = {
  children: ReactNode
}

export function ConfigProvider(props: Props) {
  const { data: config } = useSuspenseQuery({
    queryKey: queryKeys.config.all,
    queryFn: async () => {
      const res = await fetch(`${apiUrl}/config`)
      if (!res.ok) {
        throw res
      }
      return res.json() as Promise<DaytonaConfiguration>
    },
  })

  const oidcConfig: AuthProviderProps = useMemo(() => {
    const isLocalhost = window.location.hostname === 'localhost'
    const stateStore = isLocalhost ? window.sessionStorage : new InMemoryWebStorage()

    return {
      authority: config.oidc.issuer,
      client_id: config.oidc.clientId,
      extraQueryParams: {
        audience: config.oidc.audience,
      },
      scope: 'openid profile email',
      redirect_uri: window.location.origin,
      staleStateAgeInSeconds: 60,
      accessTokenExpiringNotificationTimeInSeconds: 290,
      userStore: new WebStorageStateStore({ store: stateStore }),
      onSigninCallback: (user) => {
        const state = user?.state as { returnTo?: string } | undefined
        const targetUrl = state?.returnTo || RoutePath.DASHBOARD
        window.history.replaceState({}, '', targetUrl)
        window.dispatchEvent(new PopStateEvent('popstate'))
      },
      post_logout_redirect_uri: window.location.origin,
    }
  }, [config])

  return (
    <ConfigContext.Provider value={{ ...config, apiUrl }}>
      <AuthProvider {...oidcConfig}>{props.children}</AuthProvider>
    </ConfigContext.Provider>
  )
}
