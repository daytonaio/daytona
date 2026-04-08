/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useConfig } from '@/hooks/useConfig'
import { usePrivacyConsentStore } from '@/hooks/usePrivacyConsent'
import { usePostHog } from 'posthog-js/react'
import { PostHogProvider } from 'posthog-js/react'
import { FC, ReactNode, useEffect } from 'react'

interface PostHogProviderWrapperProps {
  children: ReactNode
}

function PostHogConsentSync() {
  const posthog = usePostHog()
  const analytics = usePrivacyConsentStore((state) => state.preferences.analytics)
  const hasConsented = usePrivacyConsentStore((state) => state.hasConsented)

  useEffect(() => {
    if (!posthog || !hasConsented) return

    if (analytics) {
      posthog.opt_in_capturing()
    } else {
      posthog.opt_out_capturing()
    }
  }, [posthog, analytics, hasConsented])

  return null
}

export const PostHogProviderWrapper: FC<PostHogProviderWrapperProps> = ({ children }) => {
  const config = useConfig()

  if (!config.posthog) {
    return children
  }

  if (!config.posthog?.apiKey || !config.posthog?.host) {
    console.error('Invalid PostHog configuration')
    return children
  }

  return (
    <PostHogProvider
      apiKey={config.posthog.apiKey}
      options={{
        api_host: config.posthog.host,
        opt_out_capturing_by_default: true,
        cookieless_mode: 'on_reject',
        persistence: 'localStorage',
        person_profiles: 'always',
        autocapture: false,
        capture_pageview: false,
        capture_pageleave: true,
      }}
    >
      <PostHogConsentSync />
      {children}
    </PostHogProvider>
  )
}
