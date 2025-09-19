/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import { FacetFilter } from '@/components/ui/facet-filter'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { OrganizationUsageOverview } from '@daytonaio/api-client'
import { useMemo, useState } from 'react'
import { Bar, BarChart, CartesianGrid, ReferenceLine, XAxis, YAxis } from 'recharts'

type ResourceKey = 'cpu' | 'ram' | 'storage'

const clamp = (v: number) => Math.max(0, Math.min(100, Math.round(v * 10) / 10))

const formatDate = (value: string) => new Date(value).toLocaleDateString(undefined, { month: 'short', day: '2-digit' })

interface PastUsage {
  date: string
  peakCpuUsage: number
  peakMemoryUsage: number
  peakDiskUsage: number
}

export function LimitUsageChart({
  defaultPeriod = 30,
  defaultResources = ['ram', 'cpu', 'storage'],
  pastUsage = [],
  currentUsage,
  title,
}: {
  defaultPeriod?: number
  defaultResources?: ResourceKey[]
  pastUsage: PastUsage[]
  currentUsage?: OrganizationUsageOverview | null
  title?: React.ReactNode
}) {
  const [period, setPeriod] = useState(defaultPeriod.toString())

  const [selected, setSelected] = useState<Set<ResourceKey>>(new Set(defaultResources))

  const data = useMemo(() => {
    if (!currentUsage) {
      return []
    }

    const { totalCpuQuota, totalMemoryQuota, totalDiskQuota } = currentUsage
    return pastUsage.slice(-Number(period)).map((r) => ({
      date: r.date,
      cpu: clamp((r.peakCpuUsage / totalCpuQuota) * 100),
      ram: clamp((r.peakMemoryUsage / totalMemoryQuota) * 100),
      storage: clamp((r.peakDiskUsage / totalDiskQuota) * 100),
    }))
  }, [pastUsage, currentUsage, period])

  const config: ChartConfig = useMemo(() => {
    const full: Record<string, { label: string; color: string }> = {
      cpu: { label: 'CPU', color: 'hsl(var(--chart-3))' },
      ram: { label: 'RAM', color: 'hsl(var(--chart-2))' },
      storage: { label: 'Storage', color: 'hsl(var(--chart-1))' },
    }
    return Object.fromEntries(Object.entries(full).filter(([k]) => selected.has(k as ResourceKey))) as ChartConfig
  }, [selected])

  return (
    <div className="w-full">
      <div className="mb-3 flex items-center justify-between">
        {title}
        <div className="flex gap-2 items-center">
          <FacetFilter
            title="Resources"
            options={[
              { label: 'CPU', value: 'cpu' },
              { label: 'RAM', value: 'ram' },
              { label: 'Storage', value: 'storage' },
            ]}
            selectedValues={selected}
            setSelectedValues={(key) => setSelected(new Set(key as Set<ResourceKey>))}
          />
          <Select value={period} onValueChange={setPeriod}>
            <SelectTrigger className="w-[160px] rounded-lg" aria-label="Select the period">
              <SelectValue placeholder="Last 30 days" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value="7">Last 7 days</SelectItem>
              <SelectItem value="14">Last 14 days</SelectItem>
              <SelectItem value="30">Last 30 days</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <ChartContainer config={config} className="aspect-auto h-[300px] w-full">
        <BarChart data={data}>
          <CartesianGrid vertical={false} />
          <XAxis
            dataKey="date"
            tickLine={false}
            axisLine={false}
            tickMargin={8}
            minTickGap={32}
            tickFormatter={formatDate}
          />
          <YAxis
            domain={[0, 100]}
            tickLine={false}
            axisLine={false}
            tickMargin={8}
            tickCount={5}
            tickFormatter={(v) => `${v}%`}
          />
          <ReferenceLine y={80} strokeDasharray="6 6" stroke="#e88c3094" label="80%" />
          <ReferenceLine y={100} strokeDasharray="6 6" stroke="hsl(var(--destructive))" label="MAX" />
          <ChartTooltip
            cursor={false}
            content={
              <ChartTooltipContent
                indicator="dot"
                labelFormatter={(label) => formatDate(label)}
                valueFormatter={(value) => `${value}%`}
              />
            }
          />
          <ChartLegend content={<ChartLegendContent />} />
          {selected.has('storage') && <Bar dataKey="storage" fill="var(--color-storage)" radius={[4, 4, 0, 0]} />}
          {selected.has('ram') && <Bar dataKey="ram" fill="var(--color-ram)" radius={[4, 4, 0, 0]} />}
          {selected.has('cpu') && <Bar dataKey="cpu" fill="var(--color-cpu)" radius={[4, 4, 0, 0]} />}
        </BarChart>
      </ChartContainer>
    </div>
  )
}
