/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { FeatureFlags } from '@/enums/FeatureFlags'
import { useFeatureFlagEnabled } from 'posthog-js/react'

export const useBillingV2Enabled = (): boolean => Boolean(useFeatureFlagEnabled(FeatureFlags.BILLING_PROVIDER_V2))

// `useFeatureFlagEnabled` returns `undefined` until PostHog has loaded the flag.
// Use this when the v1/v2 choice changes the endpoint being hit, so callers can
// avoid a v1 fetch followed by a v2 refetch.
export const useBillingV2FlagLoaded = (): boolean =>
  useFeatureFlagEnabled(FeatureFlags.BILLING_PROVIDER_V2) !== undefined
