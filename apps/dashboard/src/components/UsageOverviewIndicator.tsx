/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { OrganizationUsageOverview } from '@daytonaio/api-client/src'

export const UsageOverviewIndicator = ({
  usage,
  className,
  isLive,
}: {
  usage: OrganizationUsageOverview
  className?: string
  isLive?: boolean
}) => {
  return (
    <div className={cn('flex gap-4 items-center', className)}>
      {isLive && <LiveIndicatorDot />}
      <ResourceLabel value={usage.currentCpuUsage} total={usage.totalCpuQuota} unit="vCPU" />
      <ResourceLabel value={usage.currentMemoryUsage} total={usage.totalMemoryQuota} unit="GiB" name="RAM" />
      <ResourceLabel value={usage.currentDiskUsage} total={usage.totalDiskQuota} unit="GiB" name="Storage" />
    </div>
  )
}

const ResourceLabel = ({
  value,
  total,
  unit,
  name,
}: {
  value: number
  total: number
  unit?: string
  name?: string
}) => (
  <span className="text-sm flex gap-1 items-center text-muted-foreground/70 font-mono uppercase">
    {name}
    <span className="text-foreground">{value}</span>/<span className="text-foreground">{total}</span>
    {unit}
  </span>
)

const LiveIndicatorDot = ({ className }: { className?: string }) => (
  <div className={cn('relative grid place-items-center h-2 w-2', className)}>
    <div className="animate-ping h-2 w-2 bg-green-500 rounded-full absolute " />
    <div className="h-2 w-2 bg-green-500 rounded-full " />
  </div>
)
