/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useRef } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { hasAuthParams, useAuth } from 'react-oidc-context'
import LoadingFallback from '@/components/LoadingFallback'
import { RoutePath } from '@/enums/RoutePath'

const LandingPage: React.FC = () => {
  const { activeNavigator, signinRedirect, isAuthenticated, isLoading } = useAuth()
  const location = useLocation()
  const hasTriedSigninRef = useRef(false)

  useEffect(() => {
    if (!hasAuthParams() && !isAuthenticated && !activeNavigator && !isLoading && !hasTriedSigninRef.current) {
      hasTriedSigninRef.current = true
      signinRedirect({
        state: {
          returnTo: location.pathname + location.search,
        },
      }).catch((error) => {
        console.error('Failed to start sign-in redirect:', error)
      })
    }
  }, [activeNavigator, isAuthenticated, isLoading, location.pathname, location.search, signinRedirect])

  if (isLoading) {
    return <LoadingFallback />
  }

  if (isAuthenticated) {
    return <Navigate to={`${RoutePath.DASHBOARD}${location.search}`} replace />
  }

  if (hasTriedSigninRef.current && !activeNavigator && !isLoading) {
    throw new Error('Unable to start sign-in redirect')
  }

  return <LoadingFallback />
}

export default LandingPage
