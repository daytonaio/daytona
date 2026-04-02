/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useConfig } from '@/hooks/useConfig'
import { buildInMemoryFlagConfig } from '@daytonaio/feature-flags'
import { OpenFeature, OpenFeatureProvider, InMemoryProvider } from '@openfeature/react-sdk'
import { usePostHog } from 'posthog-js/react'
import { FC, ReactNode, useRef } from 'react'
import { PostHogWebProvider } from './openfeature-posthog-web.provider'

export const OpenFeatureProviderWrapper: FC<{ children: ReactNode }> = ({ children }) => {
  const config = useConfig()
  const posthog = usePostHog()
  const initialized = useRef(false)

  if (!initialized.current) {
    if (config.posthog?.apiKey && posthog) {
      OpenFeature.setProvider(new PostHogWebProvider(posthog))
    } else {
      OpenFeature.setProvider(new InMemoryProvider(buildInMemoryFlagConfig()))
    }
    initialized.current = true
  }

  return <OpenFeatureProvider>{children}</OpenFeatureProvider>
}
