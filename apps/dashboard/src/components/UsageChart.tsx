/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import * as React from 'react'
import { Area, AreaChart, Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { subMonths } from 'date-fns'
import { useCallback, useMemo } from 'react'
import { FacetFilter } from './ui/facet-filter'

export type UsageChartData = {
  date: string
  ramGB: number
  cpu: number
  diskGB: number
  // gpu: number
}

type UsageChartProps = {
  usageData: UsageChartData[]
  showTotal?: boolean
  title?: string
}

const getShortDate = (value: string) => {
  const date = new Date(value)
  return date.toLocaleDateString('en-US', {
    month: 'short',
    year: 'numeric',
  })
}

export function UsageChart({ usageData, showTotal, title }: UsageChartProps) {
  const [timeRange, setTimeRange] = React.useState(12)
  const [chartType, setChartType] = React.useState<'bar' | 'area'>('area')
  const [filters, setFilters] = React.useState<Set<string>>(new Set(['ramGB', 'cpu', 'diskGB']))

  const filterFromObject = useCallback(
    (obj: any) => {
      return Object.fromEntries(Object.entries(obj).filter(([key]) => key === 'date' || filters.has(key)))
    },
    [filters],
  )

  const data = useMemo(() => {
    const referenceDate = new Date(Date.now())

    let monthsToSubtract = 12
    if (timeRange === 6) {
      monthsToSubtract = 6
    } else if (timeRange === 3) {
      monthsToSubtract = 3
    }
    const startDate = subMonths(new Date(referenceDate), monthsToSubtract)

    const filteredData = usageData.filter((usage) => {
      const date = new Date(usage.date)
      return date >= startDate
    })

    if (filteredData.length < timeRange) {
      const missingData: UsageChartData[] = []
      for (let i = 0; i < timeRange - filteredData.length; i++) {
        const refDate = filteredData.length > 0 ? filteredData[0].date : new Date()
        missingData.push({
          date: subMonths(new Date(refDate), i + 1).toISOString(),
          diskGB: 0,
          ramGB: 0,
          cpu: 0,
          // gpu: 0,
        })
      }
      return [...missingData.reverse(), ...filteredData].map(filterFromObject)
    }

    return filteredData.map(filterFromObject)
  }, [usageData, timeRange, filterFromObject])

  const chartConfig = useMemo(() => {
    const mapped: Record<string, { label: string; color: string }> = {
      diskGB: {
        label: 'Storage',
        color: 'hsl(var(--chart-1))',
      },
      ramGB: {
        label: 'RAM',
        color: 'hsl(var(--chart-2))',
      },
      cpu: {
        label: 'CPU',
        color: 'hsl(var(--chart-3))',
      },
      // gpu: {
      //   label: 'GPU',
      //   color: 'hsl(var(--chart-4))',
      // },
    }

    return filterFromObject(mapped)
  }, [filterFromObject])

  return (
    <Card>
      <CardHeader className="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
        <div className="grid flex-1 gap-1 text-center sm:text-left">
          <CardTitle>{title}</CardTitle>
        </div>
        <FacetFilter
          title="Filters"
          options={[
            {
              label: 'RAM',
              value: 'ramGB',
            },
            {
              label: 'CPU',
              value: 'cpu',
            },
            {
              label: 'Storage',
              value: 'diskGB',
            },
            // {
            //   label: 'GPU',
            //   value: 'gpu',
            // },
          ]}
          selectedValues={filters}
          setSelectedValues={setFilters}
        />

        <Select value={chartType} onValueChange={(value) => setChartType(value as 'bar' | 'area')}>
          <SelectTrigger className="w-[160px] rounded-lg sm:ml-auto" aria-label="Select a chart type">
            <SelectValue placeholder="Bar" />
          </SelectTrigger>
          <SelectContent className="rounded-xl">
            <SelectItem value="bar">Bar</SelectItem>
            <SelectItem value="area">Area</SelectItem>
          </SelectContent>
        </Select>
        <Select value={timeRange.toString()} onValueChange={(value) => setTimeRange(Number(value))}>
          <SelectTrigger className="w-[160px] rounded-lg sm:ml-auto" aria-label="Select the range of months">
            <SelectValue placeholder="Last 12 months" />
          </SelectTrigger>
          <SelectContent className="rounded-xl">
            <SelectItem value="12" className="rounded-lg">
              Last 12 months
            </SelectItem>
            <SelectItem value="6" className="rounded-lg">
              Last 6 months
            </SelectItem>
            <SelectItem value="3" className="rounded-lg">
              Last 3 months
            </SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        {chartType === 'bar' ? (
          <ChartContainer config={chartConfig as ChartConfig} className="aspect-auto h-[300px] w-full">
            <BarChart accessibilityLayer data={data}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="date" tickLine={false} tickMargin={10} axisLine={false} tickFormatter={getShortDate} />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickCount={4}
                tickFormatter={(value) => {
                  if (!showTotal) {
                    return value
                  }

                  return `$${value}`
                }}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    indicator="dot"
                    hideLabel={!showTotal}
                    labelFormatter={(label, payload) => {
                      return `${getShortDate(label)}: $${payload.reduce((acc, curr) => acc + (curr.value as number), 0)}`
                    }}
                  />
                }
              />
              <ChartLegend content={<ChartLegendContent />} />
              <Bar dataKey="diskGB" stackId="a" fill="var(--color-diskGB)" radius={[0, 0, 0, 0]} />
              <Bar dataKey="ramGB" stackId="a" fill="var(--color-ramGB)" radius={[0, 0, 0, 0]} />
              <Bar dataKey="cpu" stackId="a" fill="var(--color-cpu)" radius={[4, 4, 0, 0]} />
              {/* <Bar dataKey="gpu" stackId="a" fill="var(--color-gpu)" radius={[0, 0, 0, 0]} /> */}
            </BarChart>
          </ChartContainer>
        ) : (
          <ChartContainer config={chartConfig as ChartConfig} className="aspect-auto h-[300px] w-full">
            <AreaChart data={data}>
              <defs>
                <linearGradient id="fillDiskGB" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-diskGB)" stopOpacity={0.8} />
                  <stop offset="95%" stopColor="var(--color-diskGB)" stopOpacity={0.1} />
                </linearGradient>
                <linearGradient id="fillRamGB" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-ramGB)" stopOpacity={0.8} />
                  <stop offset="95%" stopColor="var(--color-ramGB)" stopOpacity={0.1} />
                </linearGradient>
                <linearGradient id="fillCpu" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-cpu)" stopOpacity={0.8} />
                  <stop offset="95%" stopColor="var(--color-cpu)" stopOpacity={0.1} />
                </linearGradient>
                {/* <linearGradient id="fillGpu" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-gpu)" stopOpacity={0.8} />
                  <stop offset="95%" stopColor="var(--color-gpu)" stopOpacity={0.1} />
                </linearGradient> */}
              </defs>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={32}
                tickFormatter={getShortDate}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickCount={4}
                tickFormatter={(value) => {
                  return `$${value}`
                }}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    indicator="dot"
                    labelFormatter={(label, payload) => {
                      return `${getShortDate(label)}: $${payload.reduce((acc, curr) => acc + (curr.value as number), 0)}`
                    }}
                  />
                }
              />
              <Area
                dataKey="diskGB"
                type="monotoneX"
                fill="url(#fillDiskGB)"
                stroke="var(--color-diskGB)"
                stackId="a"
              />
              <Area dataKey="ramGB" type="monotoneX" fill="url(#fillRamGB)" stroke="var(--color-ramGB)" stackId="b" />
              <Area dataKey="cpu" type="monotoneX" fill="url(#fillCpu)" stroke="var(--color-cpu)" stackId="c" />
              {/* <Area dataKey="gpu" type="monotoneX" fill="url(#fillGpu)" stroke="var(--color-gpu)" stackId="d" /> */}
              <ChartLegend content={<ChartLegendContent />} />
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
