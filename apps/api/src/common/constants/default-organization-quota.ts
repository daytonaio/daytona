/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateOrganizationQuotaDto } from '../../organization/dto/create-organization-quota.dto'

export const DEFAULT_ORGANIZATION_QUOTA: CreateOrganizationQuotaDto = {
  totalCpuQuota: 10,
  totalMemoryQuota: 10,
  totalDiskQuota: 30,
  maxCpuPerSandbox: 4,
  maxMemoryPerSandbox: 8,
  maxDiskPerSandbox: 10,
  snapshotQuota: 100,
  maxSnapshotSize: 20,
  volumeQuota: 100,
}
