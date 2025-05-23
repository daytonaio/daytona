/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateOrganizationQuotaDto } from '../../organization/dto/create-organization-quota.dto'

export const DEFAULT_ORGANIZATION_QUOTA: CreateOrganizationQuotaDto = {
  totalCpuQuota: 10,
  totalMemoryQuota: 10,
  totalDiskQuota: 30,
  maxCpuPerWorkspace: 4,
  maxMemoryPerWorkspace: 8,
  maxDiskPerWorkspace: 10,
  snapshotQuota: 5,
  maxSnapshotSize: 5,
  volumeQuota: 10,
}
