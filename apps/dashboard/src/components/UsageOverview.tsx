/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { SandboxClass, type RegionUsageOverview } from '@daytona/api-client'
import type { ReactNode } from 'react'
import QuotaLine from './QuotaLine'
import { Skeleton } from './ui/skeleton'

export function UsageOverview({
  usageOverview,
  hasGpuQuotaInClass = false,
  className,
}: {
  usageOverview: RegionUsageOverview
  hasGpuQuotaInClass?: boolean
  className?: string
}) {
  const isWindows = usageOverview.sandboxClass === SandboxClass.WINDOWS
  const gpuCurrent = isWindows ? 0 : usageOverview.currentGpuUsage
  const gpuTotal = isWindows ? 0 : usageOverview.totalGpuQuota

  return (
    <div className={cn('flex gap-4 [&>*]:flex-1 flex-col lg:flex-row', className)}>
      <ResourceUsageItem
        label="Compute"
        value={<UsageValue current={usageOverview.currentCpuUsage} total={usageOverview.totalCpuQuota} unit="vCPU" />}
      >
        <QuotaLine current={usageOverview.currentCpuUsage} total={usageOverview.totalCpuQuota} />
      </ResourceUsageItem>
      <ResourceUsageItem
        label="Memory"
        value={
          <UsageValue current={usageOverview.currentMemoryUsage} total={usageOverview.totalMemoryQuota} unit="GiB" />
        }
      >
        <QuotaLine current={usageOverview.currentMemoryUsage} total={usageOverview.totalMemoryQuota} />
      </ResourceUsageItem>
      <ResourceUsageItem
        label="Storage"
        value={<UsageValue current={usageOverview.currentDiskUsage} total={usageOverview.totalDiskQuota} unit="GiB" />}
      >
        <QuotaLine current={usageOverview.currentDiskUsage} total={usageOverview.totalDiskQuota} />
      </ResourceUsageItem>
      <ResourceUsageItem
        label="GPU"
        className={cn({ 'opacity-50': isWindows })}
        value={
          <UsageValue
            current={gpuCurrent}
            total={gpuTotal}
            unit="GPU"
            zeroQuotaValue={
              <GpuZeroQuotaValue isWindows={isWindows} hasGpuQuotaInClass={hasGpuQuotaInClass} current={gpuCurrent} />
            }
          />
        }
      >
        <QuotaLine current={gpuCurrent} total={gpuTotal} />
      </ResourceUsageItem>
    </div>
  )
}

function formatUsageValue(value: number) {
  const truncated = Math.trunc(value * 10) / 10

  if (Number.isInteger(truncated)) {
    return String(truncated)
  }

  return truncated.toFixed(1)
}

function UsageValue({
  current,
  total,
  unit,
  zeroQuotaValue,
}: {
  current: number
  total: number
  unit: string
  zeroQuotaValue?: ReactNode
}) {
  if (total > 0 || current > 0) {
    return <UsageLabel current={current} total={total} unit={unit} />
  }

  return zeroQuotaValue ?? <span className="text-xs text-muted-foreground text-nowrap">0 / 0 {unit}</span>
}

function GpuZeroQuotaValue({
  isWindows,
  hasGpuQuotaInClass,
  current,
}: {
  isWindows: boolean
  hasGpuQuotaInClass: boolean
  current: number
}) {
  if (current > 0) {
    return null
  }

  if (isWindows) {
    return <span className="text-xs text-muted-foreground text-nowrap">Coming soon</span>
  }

  if (hasGpuQuotaInClass) {
    return <span className="text-xs text-muted-foreground text-nowrap">Unavailable in region</span>
  }

  return (
    <a
      href="mailto:sales@daytona.io?subject=GPU%20quota%20request"
      className="text-xs font-medium text-foreground underline underline-offset-2 hover:text-muted-foreground text-nowrap"
    >
      Contact Sales
    </a>
  )
}

export function UsageOverviewSkeleton() {
  return (
    <div className="flex flex-col gap-3 p-4 lg:flex-row">
      <Skeleton className="h-8 w-full" />
      <Skeleton className="h-8 w-full" />
      <Skeleton className="h-8 w-full" />
      <Skeleton className="h-8 w-full" />
    </div>
  )
}

function ResourceUsageItem({
  label,
  value,
  className,
  children,
}: {
  label: string
  value: ReactNode
  className?: string
  children: ReactNode
}) {
  return (
    <div className={cn('flex flex-col gap-1', className)}>
      <div className="w-full flex justify-between gap-2">
        <div className="text-muted-foreground text-xs">{label}</div>
        {value}
      </div>
      {children}
    </div>
  )
}

const UsageLabel = ({ current, total, unit }: { current: number; total: number; unit: string }) => {
  const percentage = total > 0 ? (current / total) * 100 : 0
  const isHighUsage = total > 0 ? percentage > 90 : current > 0

  return (
    <span
      className={cn('text-xs text-nowrap', {
        'text-destructive-foreground': isHighUsage,
      })}
    >
      {formatUsageValue(current)} <span className="opacity-50">/</span> {formatUsageValue(total)} {unit}
    </span>
  )
}
