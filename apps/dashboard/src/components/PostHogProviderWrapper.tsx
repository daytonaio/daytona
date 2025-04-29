/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { FC, ReactNode } from 'react'
import { PostHogProvider } from 'posthog-js/react'

const posthogKey = import.meta.env.VITE_POSTHOG_KEY
const posthogHost = import.meta.env.VITE_POSTHOG_HOST

interface PostHogProviderWrapperProps {
  children: ReactNode
}

export const PostHogProviderWrapper: FC<PostHogProviderWrapperProps> = ({ children }) => {
  if (!import.meta.env.PROD) {
    return children
  }

  if (!posthogKey || !posthogHost) {
    console.error('Invalid PostHog configuration')
    return children
  }

  return (
    <PostHogProvider
      apiKey={posthogKey}
      options={{
        api_host: posthogHost,
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
