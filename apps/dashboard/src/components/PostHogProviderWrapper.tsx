/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useConfig } from '@/hooks/useConfig'
import { PostHogProvider } from 'posthog-js/react'
import { FC, ReactNode } from 'react'

interface PostHogProviderWrapperProps {
  children: ReactNode
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
      {children}
    </PostHogProvider>
  )
}
