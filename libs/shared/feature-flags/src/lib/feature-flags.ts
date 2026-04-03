/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum FeatureFlags {
  ORGANIZATION_INFRASTRUCTURE = 'organization_infrastructure',
  SANDBOX_RESIZE = 'sandbox_resize',
  ORGANIZATION_EXPERIMENTS = 'organization_experiments',
  DASHBOARD_PLAYGROUND = 'dashboard_playground',
  DASHBOARD_WEBHOOKS = 'dashboard_webhooks',
  SANDBOX_SPENDING = 'sandbox_spending',
  DASHBOARD_CREATE_SANDBOX = 'dashboard_create-sandbox',
}

// Per-flag config for @RequireFlagsEnabled decorators.
// defaultValue = fallback when the provider fails (false = block, true = allow).
export const FeatureFlagConfig: Record<keyof typeof FeatureFlags, { flagKey: FeatureFlags; defaultValue: boolean }> = {
  ORGANIZATION_INFRASTRUCTURE: { flagKey: FeatureFlags.ORGANIZATION_INFRASTRUCTURE, defaultValue: false },
  SANDBOX_RESIZE: { flagKey: FeatureFlags.SANDBOX_RESIZE, defaultValue: false },
  ORGANIZATION_EXPERIMENTS: { flagKey: FeatureFlags.ORGANIZATION_EXPERIMENTS, defaultValue: true },
  DASHBOARD_PLAYGROUND: { flagKey: FeatureFlags.DASHBOARD_PLAYGROUND, defaultValue: false },
  DASHBOARD_WEBHOOKS: { flagKey: FeatureFlags.DASHBOARD_WEBHOOKS, defaultValue: false },
  SANDBOX_SPENDING: { flagKey: FeatureFlags.SANDBOX_SPENDING, defaultValue: false },
  DASHBOARD_CREATE_SANDBOX: { flagKey: FeatureFlags.DASHBOARD_CREATE_SANDBOX, defaultValue: false },
}
