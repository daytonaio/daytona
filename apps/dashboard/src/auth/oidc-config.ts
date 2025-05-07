/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { InMemoryWebStorage, WebStorageStateStore } from 'oidc-client-ts'
import { AuthProviderProps } from 'react-oidc-context'
import { RoutePath } from '@/enums/RoutePath'

export const oidcConfig: AuthProviderProps = {
  authority: import.meta.env.VITE_OIDC_DOMAIN,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  extraQueryParams: {
    audience: import.meta.env.VITE_OIDC_AUDIENCE,
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
}
