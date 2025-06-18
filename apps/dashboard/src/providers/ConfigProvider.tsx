/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuthProvider, AuthProviderProps } from 'react-oidc-context'
import { ReactNode, useMemo } from 'react'
import { ConfigContext } from '../contexts/ConfigContext'
import { suspendedFetch } from '../lib/suspended-fetch'
import { InMemoryWebStorage, WebStorageStateStore } from 'oidc-client-ts'
import { DaytonaConfiguration } from '@daytonaio/api-client/src'
import { RoutePath } from '@/enums/RoutePath'

const getConfig = suspendedFetch<DaytonaConfiguration>(`${import.meta.env.VITE_API_URL}/config`)

type Props = {
  children: ReactNode
}

export function ConfigProvider(props: Props) {
  const config = getConfig()

  const oidcConfig: AuthProviderProps = useMemo(
    () => ({
      authority: config.oidc.issuer,
      client_id: config.oidc.clientId,
      extraQueryParams: {
        audience: config.oidc.audience,
      },
      scope: 'openid profile email offline_access',
      redirect_uri: window.location.origin,
      staleStateAgeInSeconds: 60,
      accessTokenExpiringNotificationTimeInSeconds: 290,
      userStore: new WebStorageStateStore({ store: new InMemoryWebStorage() }),
      onSigninCallback: (user) => {
        const state = user?.state as { returnTo?: string } | undefined
        const targetUrl = state?.returnTo || RoutePath.DASHBOARD
        window.history.replaceState({}, '', targetUrl)
        window.dispatchEvent(new PopStateEvent('popstate'))
      },
      post_logout_redirect_uri: window.location.origin,
    }),
    [config],
  )

  return (
    <ConfigContext.Provider value={config}>
      <AuthProvider {...oidcConfig}>{props.children}</AuthProvider>
    </ConfigContext.Provider>
  )
}
