/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { FeatureFlags } from '@/enums/FeatureFlags'
import { useFeatureFlagEnabled } from 'posthog-js/react'

export const useBillingV2Enabled = (): boolean => Boolean(useFeatureFlagEnabled(FeatureFlags.BILLING_PROVIDER_V2))
