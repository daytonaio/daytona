/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from 'react-oidc-context'
import LoadingFallback from '@/components/LoadingFallback'
import { RoutePath } from '@/enums/RoutePath'

const LandingPage: React.FC = () => {
  const { signinRedirect, isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return <LoadingFallback />
  }

  if (isAuthenticated) {
    return <Navigate to={RoutePath.Dashboard} replace />
  } else {
    void signinRedirect({
      state: {
        returnTo: window.location.pathname,
      },
    })
  }
}

export default LandingPage
