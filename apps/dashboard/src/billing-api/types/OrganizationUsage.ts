/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export interface OrganizationUsage {
  from: Date
  to: Date
  issuingDate: string
  amountCents: number
  totalAmountCents: number
  taxesAmountCents: number
  usageCharges: UsageCharge[]
}

export interface UsageCharge {
  units: string
  eventsCount: number
  amountCents: number
  billableMetric: BillableMetricCode
}

export enum BillableMetricCode {
  CPU_USAGE = 'cpu_usage',
  GPU_USAGE = 'gpu_usage',
  RAM_USAGE = 'ram_usage',
  DISK_USAGE = 'disk_usage',
  UNKNOWN = 'unknown',
}
