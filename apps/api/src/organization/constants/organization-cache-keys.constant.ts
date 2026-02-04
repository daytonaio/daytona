/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export function getOrganizationCacheKey(organizationId: string): string {
  return `organization:${organizationId}`
}

export function getOrganizationUserCacheKey(organizationId: string, userId: string): string {
  return `organization-user:${organizationId}:${userId}`
}
