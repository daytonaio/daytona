/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useFeatureFlagEnabled as phFfEnabled } from 'posthog-js/react'
import { useConfig } from './useConfig'
import { FeatureFlags } from '@/enums/FeatureFlags'

export function useFeatureFlagEnabled(flag: FeatureFlags): boolean {
  const config = useConfig()
  const isFlagEnabled = phFfEnabled(flag)

  if (!config.posthog) {
    return true
  }

  return !!isFlagEnabled
}
