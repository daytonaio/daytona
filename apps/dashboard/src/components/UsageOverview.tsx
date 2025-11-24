/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { OrganizationUsageOverview } from '@daytonaio/api-client'
import { AlertTriangle, CpuIcon, HardDriveIcon, MemoryStickIcon } from 'lucide-react'
import QuotaLine from './QuotaLine'
import { Skeleton } from './ui/skeleton'

export function UsageOverview({ usageOverview }: { usageOverview: OrganizationUsageOverview }) {
  return (
    <div className="flex gap-4 flex-col">
      <div className="flex flex-col gap-1">
        <div className="w-full flex justify-between">
          <div className="flex items-center gap-2 text-muted-foreground">
            <CpuIcon size={16} className="opacity-50" /> Compute
          </div>
          <UsageLabel current={usageOverview.currentCpuUsage} total={usageOverview.totalCpuQuota} unit="vCPU" />
        </div>
        <QuotaLine current={usageOverview.currentCpuUsage} total={usageOverview.totalCpuQuota} />
      </div>
      <div className="flex flex-col gap-1">
        <div className="w-full flex justify-between">
          <div className="flex items-center gap-2 text-muted-foreground">
            <MemoryStickIcon size={16} className="opacity-50" /> Memory
          </div>
          <UsageLabel current={usageOverview.currentMemoryUsage} total={usageOverview.totalMemoryQuota} unit="GiB" />
        </div>
        <QuotaLine current={usageOverview.currentMemoryUsage} total={usageOverview.totalMemoryQuota} />
      </div>
      <div className="flex flex-col gap-1">
        <div className="w-full flex justify-between">
          <div className="flex items-center gap-2 text-muted-foreground">
            <HardDriveIcon size={16} className="opacity-50" /> Storage
          </div>
          <UsageLabel current={usageOverview.currentDiskUsage} total={usageOverview.totalDiskQuota} unit="GiB" />
        </div>
        <QuotaLine current={usageOverview.currentDiskUsage} total={usageOverview.totalDiskQuota} />
      </div>
    </div>
  )
}

export function UsageOverviewSkeleton() {
  return (
    <div className="flex flex-col gap-4">
      <Skeleton className="h-9 w-full" />
      <Skeleton className="h-9 w-full" />
      <Skeleton className="h-9 w-full" />
    </div>
  )
}

const UsageLabel = ({ current, total, unit }: { current: number; total: number; unit: string }) => {
  const percentage = (current / total) * 100
  const isHighUsage = percentage > 90

  return (
    <div className="flex items-center gap-1">
      {isHighUsage && <AlertTriangle className="w-4 h-4 text-red-500" />}
      <span
        className={cn('font-mono', {
          'text-destructive': isHighUsage,
        })}
      >
        {current} <span className="opacity-50">/</span> {total} {unit}
      </span>
    </div>
  )
}
