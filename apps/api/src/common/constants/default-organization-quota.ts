/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateOrganizationQuotaDto } from '../../organization/dto/create-organization-quota.dto'

export const DEFAULT_ORGANIZATION_QUOTA: CreateOrganizationQuotaDto = {
  totalCpuQuota: 20,
  totalMemoryQuota: 40,
  totalDiskQuota: 50,
  maxCpuPerWorkspace: 4,
  maxMemoryPerWorkspace: 8,
  maxDiskPerWorkspace: 10,
  maxConcurrentWorkspaces: 10,
  workspaceQuota: 10,
  imageQuota: 5,
  maxImageSize: 5,
  totalImageSize: 10,
}
