/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BillableMetricCode, OrganizationUsage } from '@/billing-api/types/OrganizationUsage'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { AggregatedUsageChart, ResourceUsageBreakdown, UsageSummary } from '@/components/spending/AggregatedUsageChart'
import { CostBreakdown } from '@/components/spending/CostBreakdown'
import { UsageChartData } from '@/components/spending/ResourceUsageChart'
import { SandboxUsageTable } from '@/components/spending/SandboxUsageTable'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle } from '@/components/ui/card'
import { DateRangePicker, QuickRangesConfig } from '@/components/ui/date-range-picker'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Separator } from '@/components/ui/separator'
import { FeatureFlags } from '@daytonaio/feature-flags'
import { UsageTimelineChart } from '@/components/spending/UsageTimelineChart'
import { useAggregatedUsage, useSandboxesUsage, useUsageChart } from '@/hooks/queries/useAnalyticsUsage'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useOrganizationUsageQuery } from '@/hooks/queries/useOrganizationUsageQuery'
import { usePastOrganizationUsageQuery } from '@/hooks/queries/usePastOrganizationUsageQuery'
import { useConfig } from '@/hooks/useConfig'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { addDays, differenceInCalendarDays, subDays } from 'date-fns'
import { AlertCircle, BarChart3, RefreshCw } from 'lucide-react'
import { useBooleanFlagValue } from '@openfeature/react-sdk'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { DateRange } from 'react-day-picker'

const analyticsQuickRanges: QuickRangesConfig = {
  hours: [1, 6, 12, 24],
  days: [7, 14, 30],
}

const Spending = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const config = useConfig()
  const spendingEnabled = useBooleanFlagValue(FeatureFlags.SANDBOX_SPENDING, false)
  const analyticsAvailable = spendingEnabled && !!config.analyticsApiUrl

  const [analyticsDateRange, setAnalyticsDateRange] = useState<DateRange>(() => {
    const now = new Date()
    return { from: subDays(now, 30), to: now }
  })

  const handleAnalyticsDateRangeChange = useCallback((range: DateRange) => {
    if (range.from && range.to) {
      const days = differenceInCalendarDays(range.to, range.from)
      if (days > 30) {
        setAnalyticsDateRange({ from: range.from, to: addDays(range.from, 30) })
        return
      }
    }
    setAnalyticsDateRange(range)
  }, [])

  const [selectedChartRegion, setSelectedChartRegion] = useState<string | undefined>(undefined)
  const hasDefaultedRegion = useRef(false)

  const analyticsParams = {
    from: analyticsDateRange.from ?? subDays(new Date(), 30),
    to: analyticsDateRange.to ?? new Date(),
    enabled: analyticsAvailable && !!selectedOrganization,
  }

  const {
    data: aggregatedUsage,
    isLoading: aggregatedLoading,
    isError: aggregatedError,
    refetch: refetchAggregated,
  } = useAggregatedUsage(analyticsParams)
  const {
    data: sandboxesUsage,
    isLoading: sandboxesLoading,
    isError: sandboxesError,
    refetch: refetchSandboxes,
  } = useSandboxesUsage(analyticsParams)
  const { data: usageChartPoints, isLoading: chartLoading } = useUsageChart({
    ...analyticsParams,
    region: selectedChartRegion,
  })

  const { data: usageOverview } = useOrganizationUsageOverviewQuery({
    organizationId: selectedOrganization?.id ?? '',
  })

  // Default chart region to the organization's default region (only once)
  useEffect(() => {
    if (hasDefaultedRegion.current) return
    const regionUsage = usageOverview?.regionUsage
    if (!regionUsage?.length) return
    hasDefaultedRegion.current = true
    const defaultRegionId = selectedOrganization?.defaultRegionId
    if (defaultRegionId && regionUsage.some((r) => r.regionId === defaultRegionId)) {
      setSelectedChartRegion(defaultRegionId)
    } else {
      setSelectedChartRegion(regionUsage[0].regionId)
    }
  }, [usageOverview?.regionUsage, selectedOrganization?.defaultRegionId])

  const {
    data: currentOrganizationUsage,
    isLoading: currentUsageLoading,
    isError: currentUsageError,
    refetch: refetchCurrentUsage,
  } = useOrganizationUsageQuery({
    organizationId: selectedOrganization?.id ?? '',
    enabled: !!selectedOrganization,
  })

  const {
    data: pastOrganizationUsage,
    isLoading: pastUsageLoading,
    isError: pastUsageError,
    refetch: refetchPastUsage,
  } = usePastOrganizationUsageQuery({
    organizationId: selectedOrganization?.id ?? '',
    enabled: !!selectedOrganization,
  })

  const sortedPastUsage = useMemo(
    () => [...(pastOrganizationUsage ?? [])].sort((a, b) => new Date(a.from).getTime() - new Date(b.from).getTime()),
    [pastOrganizationUsage],
  )

  const usageChartData = useMemo(
    () =>
      [...sortedPastUsage, ...(currentOrganizationUsage ? [currentOrganizationUsage] : [])].map(
        convertUsageToChartData,
      ),
    [sortedPastUsage, currentOrganizationUsage],
  )

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Spending</PageTitle>
      </PageHeader>

      <PageContent>
        {analyticsAvailable && (
          <Card>
            <CardHeader className="flex flex-row items-center gap-2 space-y-0 border-b p-4">
              <div className="flex-1">
                <CardTitle>Resource Usage</CardTitle>
              </div>
              <DateRangePicker
                value={analyticsDateRange}
                onChange={handleAnalyticsDateRangeChange}
                quickRangesEnabled
                quickRanges={analyticsQuickRanges}
                timeSelection
                defaultSelectedQuickRange="Last 30 days"
                className="w-auto"
                contentAlign="end"
              />
            </CardHeader>
            {aggregatedError ? (
              <Empty className="py-12">
                <EmptyHeader>
                  <EmptyMedia variant="icon" className="bg-destructive-background text-destructive">
                    <AlertCircle />
                  </EmptyMedia>
                  <EmptyTitle className="text-destructive">Failed to load resource usage</EmptyTitle>
                  <EmptyDescription>Something went wrong while fetching usage data. Please try again.</EmptyDescription>
                </EmptyHeader>
                <EmptyContent>
                  <Button variant="secondary" size="sm" onClick={() => refetchAggregated()}>
                    <RefreshCw />
                    Retry
                  </Button>
                </EmptyContent>
              </Empty>
            ) : !aggregatedLoading && !aggregatedUsage?.sandboxCount ? (
              <Empty className="py-12">
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <BarChart3 />
                  </EmptyMedia>
                  <EmptyTitle>No resource usage data</EmptyTitle>
                  <EmptyDescription>
                    Usage data will appear here once your sandboxes start consuming resources in the selected time
                    range.
                  </EmptyDescription>
                </EmptyHeader>
              </Empty>
            ) : (
              <>
                <UsageSummary data={aggregatedUsage} isLoading={aggregatedLoading} />
                <Separator />
                <AggregatedUsageChart data={aggregatedUsage} isLoading={aggregatedLoading} />
                <Separator />
                <ResourceUsageBreakdown data={aggregatedUsage} />
                <Separator />
                <UsageTimelineChart
                  data={usageChartPoints}
                  isLoading={chartLoading}
                  regionUsage={usageOverview?.regionUsage}
                  selectedRegion={selectedChartRegion}
                  onRegionChange={setSelectedChartRegion}
                  dateRange={{ from: analyticsParams.from, to: analyticsParams.to }}
                />
              </>
            )}
            <Separator />
            <div className="p-4">
              <p className="text-xl font-semibold leading-none tracking-tight">Per-Sandbox Usage</p>
              <p className="text-sm text-muted-foreground mt-2">
                Resource consumption broken down by individual sandbox.
              </p>
            </div>
            {sandboxesError ? (
              <Empty className="py-12">
                <EmptyHeader>
                  <EmptyMedia variant="icon" className="bg-destructive-background text-destructive">
                    <AlertCircle />
                  </EmptyMedia>
                  <EmptyTitle className="text-destructive">Failed to load sandbox usage</EmptyTitle>
                  <EmptyDescription>
                    Something went wrong while fetching per-sandbox data. Please try again.
                  </EmptyDescription>
                </EmptyHeader>
                <EmptyContent>
                  <Button variant="secondary" size="sm" onClick={() => refetchSandboxes()}>
                    <RefreshCw />
                    Retry
                  </Button>
                </EmptyContent>
              </Empty>
            ) : !sandboxesLoading && !sandboxesUsage?.length ? (
              <Empty className="py-12">
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <BarChart3 />
                  </EmptyMedia>
                  <EmptyTitle>No sandbox usage yet</EmptyTitle>
                  <EmptyDescription>
                    Once you create and run a sandbox, its resource consumption will appear here.
                  </EmptyDescription>
                </EmptyHeader>
              </Empty>
            ) : (
              <SandboxUsageTable data={sandboxesUsage} isLoading={sandboxesLoading} />
            )}
          </Card>
        )}

        <CostBreakdown
          usageData={usageChartData}
          showTotal
          isLoading={currentUsageLoading || pastUsageLoading}
          isError={currentUsageError || pastUsageError}
          onRetry={() => {
            if (currentUsageError) refetchCurrentUsage()
            if (pastUsageError) refetchPastUsage()
          }}
        />
      </PageContent>
    </PageLayout>
  )
}

function convertUsageToChartData(usage: OrganizationUsage): UsageChartData {
  let ramGB = 0
  let cpu = 0
  let diskGB = 0
  // let gpu = 0

  for (const charge of usage.usageCharges) {
    switch (charge.billableMetric) {
      case BillableMetricCode.RAM_USAGE:
        ramGB += Number(charge.amountCents) / 100
        break
      case BillableMetricCode.CPU_USAGE:
        cpu += Number(charge.amountCents) / 100
        break
      case BillableMetricCode.DISK_USAGE:
        diskGB += Number(charge.amountCents) / 100
        break
      // case BillableMetricCode.GPU_USAGE:
      //   gpu += Number(charge.amountCents) / 100
      //   break
    }
  }

  return {
    date: new Date(usage.from).toISOString(),
    diskGB,
    ramGB,
    cpu,
    // gpu,
  }
}

export default Spending
