/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { FC, ReactNode } from 'react'
import { PostHogProvider } from 'posthog-js/react'
import { useConfig } from '@/hooks/useConfig'

interface PostHogProviderWrapperProps {
  children: ReactNode
}

export const PostHogProviderWrapper: FC<PostHogProviderWrapperProps> = ({ children }) => {
  const config = useConfig()

  if (!import.meta.env.PROD || !config.posthog) {
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
        person_profiles: 'always',
        autocapture: false, // ignore default frontend events
        capture_pageview: false, // initial pageview (handled in App.tsx)
        capture_pageleave: true, // end of session
      }}
    >
      {children}
    </PostHogProvider>
  )
}
