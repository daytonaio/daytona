/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiContext } from '@/contexts/ApiContext'
import { useEffect, useRef, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import LoadingFallback from '@/components/LoadingFallback'
import { ApiClient } from '@/api/apiClient'
import { useLocation } from 'react-router-dom'
import { useConfig } from '@/hooks/useConfig'

export const ApiProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user, isAuthenticated, isLoading, signinRedirect } = useAuth()
  const config = useConfig()
  const location = useLocation()

  const apiRef = useRef<ApiClient | null>(null)
  const [isApiReady, setIsApiReady] = useState(false)

  // Initialize API client as soon as user is available
  useEffect(() => {
    if (user) {
      if (!apiRef.current) {
        apiRef.current = new ApiClient(config, user.access_token)
      } else {
        apiRef.current.setAccessToken(user.access_token)
      }
      setIsApiReady(true)
    } else {
      setIsApiReady(false)
    }
  }, [user, config])

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      void signinRedirect({
        state: {
          returnTo: location.pathname + location.search,
        },
      })
    }
  }, [isLoading, isAuthenticated, signinRedirect, location])

  if (isLoading || !isApiReady) {
    return <LoadingFallback />
  }

  return <ApiContext.Provider value={apiRef.current}>{children}</ApiContext.Provider>
}
