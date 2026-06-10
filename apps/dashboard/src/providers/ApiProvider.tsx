/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiContext } from '@/contexts/ApiContext'
import { useEffect, useMemo, useRef, useState } from 'react'
import { hasAuthParams, useAuth } from 'react-oidc-context'
import LoadingFallback from '@/components/LoadingFallback'
import { ApiClient } from '@/api/apiClient'
import { useLocation } from 'react-router-dom'
import { useConfig } from '@/hooks/useConfig'

export const ApiProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { activeNavigator, user, isAuthenticated, isLoading, signinRedirect } = useAuth()
  const config = useConfig()
  const location = useLocation()

  const hasTriedSigninRef = useRef(false)
  const [hasTriedSignin, setHasTriedSignin] = useState(false)

  const api = useMemo(() => {
    if (!isAuthenticated || !user) {
      return null
    }

    return new ApiClient(config, user.access_token)
  }, [config, isAuthenticated, user?.access_token])

  useEffect(() => {
    if (!hasAuthParams() && !isAuthenticated && !activeNavigator && !isLoading && !hasTriedSigninRef.current) {
      hasTriedSigninRef.current = true
      setHasTriedSignin(true)
      signinRedirect({
        state: {
          returnTo: location.pathname + location.search,
        },
      }).catch((error) => {
        console.error('Failed to start sign-in redirect:', error)
      })
    }
  }, [activeNavigator, isAuthenticated, isLoading, location.pathname, location.search, signinRedirect])

  if (hasTriedSignin && !isAuthenticated && !activeNavigator && !isLoading) {
    throw new Error('Unable to start sign-in redirect')
  }

  if (isLoading || !api) {
    let loadingSource = 'api-provider:missing-api'
    if (activeNavigator) {
      loadingSource = `api-provider:${activeNavigator}`
    } else if (isLoading) {
      loadingSource = 'api-provider:auth-loading'
    }

    return <LoadingFallback source={loadingSource} />
  }

  return <ApiContext.Provider value={api}>{children}</ApiContext.Provider>
}
