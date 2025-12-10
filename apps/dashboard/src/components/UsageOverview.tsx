/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { RegionUsageOverview } from '@daytonaio/api-client'
import QuotaLine from './QuotaLine'
import { Skeleton } from './ui/skeleton'

export function UsageOverview({
  usageOverview,
  className,
}: {
  usageOverview: RegionUsageOverview
  className?: string
}) {
  return (
    <div className={cn('flex gap-4 [&>*]:flex-1 flex-col lg:flex-row', className)}>
      <div className="flex flex-col gap-1">
        <div className="w-full flex justify-between gap-2">
          <div className="text-muted-foreground text-xs">Compute</div>
          <UsageLabel current={usageOverview.currentCpuUsage} total={usageOverview.totalCpuQuota} unit="vCPU" />
        </div>
        <QuotaLine current={usageOverview.currentCpuUsage} total={usageOverview.totalCpuQuota} />
      </div>
      <div className="flex flex-col gap-1">
        <div className="w-full flex justify-between gap-2">
          <div className="text-muted-foreground text-xs">Memory</div>
          <UsageLabel current={usageOverview.currentMemoryUsage} total={usageOverview.totalMemoryQuota} unit="GiB" />
        </div>
        <QuotaLine current={usageOverview.currentMemoryUsage} total={usageOverview.totalMemoryQuota} />
      </div>
      <div className="flex flex-col gap-1">
        <div className="w-full flex justify-between gap-2">
          <div className="text-muted-foreground text-xs">Storage</div>
          <UsageLabel current={usageOverview.currentDiskUsage} total={usageOverview.totalDiskQuota} unit="GiB" />
        </div>
        <QuotaLine current={usageOverview.currentDiskUsage} total={usageOverview.totalDiskQuota} />
      </div>
    </div>
  )
}

export function UsageOverviewSkeleton() {
  return (
    <div className="flex flex-col gap-3 p-4 lg:flex-row">
      <Skeleton className="h-8 w-full" />
      <Skeleton className="h-8 w-full" />
      <Skeleton className="h-8 w-full" />
    </div>
  )
}

const UsageLabel = ({ current, total, unit }: { current: number; total: number; unit: string }) => {
  const percentage = (current / total) * 100
  const isHighUsage = percentage > 90

  return (
    <span
      className={cn('text-xs text-nowrap', {
        'text-destructive': isHighUsage,
      })}
    >
      {current} <span className="opacity-50">/</span> {total} {unit}
    </span>
  )
}
