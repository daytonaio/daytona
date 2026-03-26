/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import LoadingFallback from '@/components/LoadingFallback'
import { usePostHog } from 'posthog-js/react'
import { useEffect } from 'react'
import { useAuth } from 'react-oidc-context'

const Logout = () => {
  const { signoutRedirect } = useAuth()
  const posthog = usePostHog()

  useEffect(() => {
    posthog?.reset()
    void signoutRedirect()
  }, [signoutRedirect, posthog])

  return <LoadingFallback />
}

export default Logout
