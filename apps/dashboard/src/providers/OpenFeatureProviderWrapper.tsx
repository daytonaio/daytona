/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useConfig } from '@/hooks/useConfig'
import { buildInMemoryFlagConfig } from '@daytonaio/feature-flags'
import { OpenFeature, OpenFeatureProvider, InMemoryProvider } from '@openfeature/react-sdk'
import { usePostHog } from 'posthog-js/react'
import { FC, ReactNode, useEffect } from 'react'
import { PostHogWebProvider } from './openfeature-posthog-web.provider'

export const OpenFeatureProviderWrapper: FC<{ children: ReactNode }> = ({ children }) => {
  const config = useConfig()
  const posthog = usePostHog()

  useEffect(() => {
    if (config.posthog?.apiKey && posthog) {
      OpenFeature.setProvider(new PostHogWebProvider(posthog))
    } else if (import.meta.env.DEV) {
      OpenFeature.setProvider(new InMemoryProvider(buildInMemoryFlagConfig()))
    } else {
      OpenFeature.setProvider(new InMemoryProvider({}))
    }
  }, [config.posthog?.apiKey, posthog])

  return <OpenFeatureProvider>{children}</OpenFeatureProvider>
}
