/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateOrganizationQuotaDto } from '../../organization/dto/create-organization-quota.dto'

export const DEFAULT_ORGANIZATION_QUOTA: CreateOrganizationQuotaDto = {
  totalCpuQuota: 10,
  totalMemoryQuota: 10,
  totalDiskQuota: 30,
  maxCpuPerWorkspace: 2,
  maxMemoryPerWorkspace: 2,
  maxDiskPerWorkspace: 3,
  maxConcurrentWorkspaces: 10,
  workspaceQuota: 20,
  imageQuota: 3,
  maxImageSize: 3,
  totalImageSize: 3,
  volumeQuota: 3,
}
