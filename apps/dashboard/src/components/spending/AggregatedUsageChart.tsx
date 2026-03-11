/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { ModelsAggregatedUsage } from '@daytonaio/analytics-api-client'
import NumberFlow from '@number-flow/react'
import { motion } from 'framer-motion'
import React from 'react'

interface AggregatedUsageChartProps {
  data: ModelsAggregatedUsage | undefined
  isLoading: boolean
}

function formatSeconds(seconds: number): { value: number; suffix: string } {
  if (seconds < 60) return { value: Math.round(seconds * 10) / 10, suffix: ' s' }
  if (seconds < 3600) return { value: Math.round((seconds / 60) * 10) / 10, suffix: ' m' }
  return { value: Math.round((seconds / 3600) * 10) / 10, suffix: ' h' }
}

function formatGBSeconds(gbSeconds: number): { value: number; suffix: string } {
  if (gbSeconds < 3600) return { value: Math.round(gbSeconds * 10) / 10, suffix: ' GB-s' }
  return { value: Math.round((gbSeconds / 3600) * 10) / 10, suffix: ' GB-h' }
}

const transition = {
  type: 'spring',
  stiffness: 60,
  damping: 15,
  mass: 1,
} as const

const SEGMENTS = [
  { key: 'cpu' as const, label: 'CPU', color: 'bg-[hsl(var(--chart-1))]' },
  { key: 'ram' as const, label: 'RAM', color: 'bg-[hsl(var(--chart-2))]' },
  { key: 'disk' as const, label: 'Disk', color: 'bg-[hsl(var(--chart-3))]' },
]

export const UsageSummary: React.FC<AggregatedUsageChartProps> = ({ data, isLoading }) => {
  const totalPrice = data?.totalPrice ?? 0
  const sandboxCount = data?.sandboxCount ?? 0

  return (
    <div className="flex gap-4 sm:gap-12 sm:flex-row flex-col p-4">
      <div className="flex flex-col gap-1">
        <div>Total Cost</div>
        <div className="relative">
          <div className={cn('text-2xl font-semibold', isLoading && 'invisible')}>
            $
            <NumberFlow
              value={Math.round(totalPrice * 100) / 100}
              format={{ minimumFractionDigits: 2, maximumFractionDigits: 2 }}
            />
          </div>
          {isLoading && <Skeleton className="absolute inset-y-1 left-0 w-24" />}
        </div>
      </div>
      <div className="flex flex-col gap-1">
        <div>Sandboxes</div>
        <div className="relative">
          <div className={cn('text-2xl font-semibold', isLoading && 'invisible')}>
            <NumberFlow value={sandboxCount} />
          </div>
          {isLoading && <Skeleton className="absolute inset-y-1 left-0 w-14" />}
        </div>
      </div>
    </div>
  )
}

export const AggregatedUsageChart: React.FC<AggregatedUsageChartProps> = ({ data, isLoading }) => {
  const cpuSeconds = data?.totalCPUSeconds ?? 0
  const ramGBSeconds = data?.totalRAMGBSeconds ?? 0
  const diskGBSeconds = data?.totalDiskGBSeconds ?? 0

  const cpu = formatSeconds(cpuSeconds)
  const ram = formatGBSeconds(ramGBSeconds)
  const disk = formatGBSeconds(diskGBSeconds)

  return (
    <div className="overflow-hidden">
      <div className="grid grid-cols-1 sm:grid-cols-3 -mr-px -mb-px">
        <StatItem label="CPU" suffix={cpu.suffix} isLoading={isLoading}>
          <NumberFlow value={cpu.value} format={{ minimumFractionDigits: 1, maximumFractionDigits: 1 }} />
        </StatItem>
        <StatItem label="RAM" suffix={ram.suffix} isLoading={isLoading}>
          <NumberFlow value={ram.value} format={{ minimumFractionDigits: 1, maximumFractionDigits: 1 }} />
        </StatItem>
        <StatItem label="Disk" suffix={disk.suffix} isLoading={isLoading}>
          <NumberFlow value={disk.value} format={{ minimumFractionDigits: 1, maximumFractionDigits: 1 }} />
        </StatItem>
      </div>
    </div>
  )
}

function StatItem({
  label,
  suffix,
  children,
  isLoading,
}: {
  label: string
  suffix?: string
  children: React.ReactNode
  isLoading?: boolean
}) {
  return (
    <div className="px-4 py-2 sm:p-4 border-b border-r border-border flex items-center gap-2 sm:block">
      <p className="text-sm text-muted-foreground">{label}</p>
      <div className="relative">
        <p className={cn('text-xl font-semibold', isLoading && 'invisible')}>
          {children}
          {suffix && <span className="text-base font-medium text-muted-foreground">{suffix}</span>}
        </p>
        {isLoading && <Skeleton className="absolute top-1 h-5 w-16" />}
      </div>
    </div>
  )
}

export const ResourceUsageBreakdown: React.FC<{ data: ModelsAggregatedUsage | undefined }> = ({ data }) => {
  const cpuSeconds = data?.totalCPUSeconds ?? 0
  const ramGBSeconds = data?.totalRAMGBSeconds ?? 0
  const diskGBSeconds = data?.totalDiskGBSeconds ?? 0

  const segmentValues = {
    cpu: cpuSeconds,
    ram: ramGBSeconds,
    disk: diskGBSeconds,
  }
  const total = cpuSeconds + ramGBSeconds + diskGBSeconds

  return (
    <div className="flex flex-col gap-4 p-4">
      <p className="text-xl font-semibold leading-none tracking-tight">Resource Breakdown</p>
      <div className="flex flex-col gap-2">
        <div className="w-full h-2 bg-muted rounded-full overflow-clip flex">
          {total === 0
            ? null
            : SEGMENTS.map(({ key, color }) => {
                const value = segmentValues[key]
                if (value === 0) return null
                const pct = (value / total) * 100
                return (
                  <motion.div
                    key={key}
                    className={cn('h-full', color)}
                    initial={{ width: 0 }}
                    animate={{ width: `${pct}%` }}
                    transition={transition}
                  />
                )
              })}
        </div>
        <div className="flex items-center gap-4 flex-wrap">
          {SEGMENTS.map(({ key, label, color }) => (
            <div key={key} className="flex items-center gap-1.5 text-xs">
              <div className={cn('w-2 h-2 rounded-full', color)} />
              <span>
                {label} {total > 0 ? `${Math.round((segmentValues[key] / total) * 100)}%` : '0%'}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
