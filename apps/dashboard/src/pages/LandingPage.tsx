/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from 'react-oidc-context'
import LoadingFallback from '@/components/LoadingFallback'
import { usePostHog } from 'posthog-js/react'

const LandingPage: React.FC = () => {
  const { signinRedirect, isAuthenticated, isLoading, user } = useAuth()
  const posthog = usePostHog()

  useEffect(() => {
    if (import.meta.env.PROD && isAuthenticated && user && posthog?.get_distinct_id() !== user.profile.sub) {
      posthog?.identify(user.profile.sub, {
        email: user.profile.email,
        name: user.profile.name,
      })
    }
  }, [isAuthenticated, user, posthog])

  if (isLoading) {
    return <LoadingFallback />
  }

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  } else {
    void signinRedirect({
      state: {
        returnTo: window.location.pathname,
      },
    })
  }
}

export default LandingPage
