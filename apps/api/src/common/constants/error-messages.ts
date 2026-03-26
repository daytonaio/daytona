/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const UPGRADE_TIER_MESSAGE = (dashboardUrl: string) =>
  `To increase concurrency limits, upgrade your organization's Tier by visiting ${dashboardUrl}/limits.`

export const ARCHIVE_SANDBOXES_MESSAGE = 'Consider archiving your unused Sandboxes to free up available storage.'

export const PER_SANDBOX_LIMIT_MESSAGE =
  'Need higher resource limits per-sandbox? Contact us at support@daytona.io and let us know about your use case.'
