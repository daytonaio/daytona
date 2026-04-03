/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { FeatureFlags } from './feature-flags'

export const DevFeatureFlags: Record<FeatureFlags, boolean> = {
  [FeatureFlags.ORGANIZATION_INFRASTRUCTURE]: true,
  [FeatureFlags.SANDBOX_RESIZE]: true,
  [FeatureFlags.ORGANIZATION_EXPERIMENTS]: true,
  [FeatureFlags.DASHBOARD_PLAYGROUND]: true,
  [FeatureFlags.DASHBOARD_WEBHOOKS]: true,
  [FeatureFlags.SANDBOX_SPENDING]: true,
  [FeatureFlags.DASHBOARD_CREATE_SANDBOX]: true,
}

export function buildInMemoryFlagConfig(
  flagValues: Record<string, boolean> = DevFeatureFlags,
): Record<string, { variants: Record<string, boolean>; disabled: boolean; defaultVariant: string }> {
  return Object.fromEntries(
    Object.entries(flagValues).map(([key, enabled]) => [
      key,
      {
        variants: { on: true, off: false },
        disabled: false,
        defaultVariant: enabled ? 'on' : 'off',
      },
    ]),
  )
}
