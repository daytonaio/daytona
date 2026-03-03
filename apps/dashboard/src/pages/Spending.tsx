/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BillableMetricCode, OrganizationUsage } from '@/billing-api/types/OrganizationUsage'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { AggregatedUsageChart, ResourceUsageBreakdown, UsageSummary } from '@/components/spending/AggregatedUsageChart'
import { SandboxUsageTable } from '@/components/spending/SandboxUsageTable'
import { UsageChart, UsageChartData } from '@/components/spending/UsageChart'
import { Card, CardHeader, CardTitle } from '@/components/ui/card'
import { DateRangePicker, QuickRangesConfig } from '@/components/ui/date-range-picker'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { FeatureFlags } from '@/enums/FeatureFlags'
import { useAggregatedUsage, useSandboxesUsage } from '@/hooks/queries/useAnalyticsUsage'
import { useOrganizationUsageQuery } from '@/hooks/queries/useOrganizationUsageQuery'
import { usePastOrganizationUsageQuery } from '@/hooks/queries/usePastOrganizationUsageQuery'
import { useConfig } from '@/hooks/useConfig'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { subDays } from 'date-fns'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import { useMemo, useState } from 'react'
import { DateRange } from 'react-day-picker'

const analyticsQuickRanges: QuickRangesConfig = {
  hours: [1, 6, 12, 24],
  days: [7, 14, 30],
}

const SKELETON_ROWS = 10

const Spending = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const config = useConfig()
  const spendingEnabled = useFeatureFlagEnabled(FeatureFlags.SANDBOX_SPENDING)
  const analyticsAvailable = spendingEnabled && !!config.analyticsApiUrl

  const [analyticsDateRange, setAnalyticsDateRange] = useState<DateRange>(() => {
    const now = new Date()
    return { from: subDays(now, 30), to: now }
  })

  const analyticsParams = {
    from: analyticsDateRange.from ?? subDays(new Date(), 30),
    to: analyticsDateRange.to ?? new Date(),
    enabled: analyticsAvailable && !!selectedOrganization,
  }

  const { data: aggregatedUsage, isLoading: aggregatedLoading } = useAggregatedUsage(analyticsParams)
  const { data: sandboxesUsage, isLoading: sandboxesLoading } = useSandboxesUsage(analyticsParams)

  const { data: currentOrganizationUsage } = useOrganizationUsageQuery({
    organizationId: selectedOrganization?.id ?? '',
    enabled: !!selectedOrganization,
  })

  const { data: pastOrganizationUsage } = usePastOrganizationUsageQuery({
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
                onChange={setAnalyticsDateRange}
                quickRangesEnabled
                quickRanges={analyticsQuickRanges}
                timeSelection
                defaultSelectedQuickRange="Last 30 days"
                className="w-auto"
                contentAlign="end"
              />
            </CardHeader>
            <UsageSummary data={aggregatedUsage} isLoading={aggregatedLoading} />
            <Separator />
            <AggregatedUsageChart data={aggregatedUsage} isLoading={aggregatedLoading} />
            <Separator />
            <ResourceUsageBreakdown data={aggregatedUsage} />
            <Separator />
            <div className="p-4">
              <p className="text-xl font-semibold leading-none tracking-tight">Per-Sandbox Usage</p>
              <p className="text-sm text-muted-foreground mt-2">
                Resource consumption broken down by individual sandbox.
              </p>
            </div>
            {sandboxesLoading ? (
              <SandboxUsageTableSkeleton />
            ) : (
              <SandboxUsageTable data={sandboxesUsage} isLoading={sandboxesLoading} />
            )}
          </Card>
        )}

        <UsageChart title="Monthly Cost Breakdown" usageData={usageChartData} showTotal />
      </PageContent>
    </PageLayout>
  )
}

function SandboxUsageTableSkeleton() {
  return (
    <div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Sandbox ID</TableHead>
            <TableHead className="text-right">Total Price</TableHead>
            <TableHead className="text-right">CPU (seconds)</TableHead>
            <TableHead className="text-right">RAM (GB-seconds)</TableHead>
            <TableHead className="text-right">Disk (GB-seconds)</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: SKELETON_ROWS }).map((_, i) => (
            <TableRow key={i}>
              <TableCell>
                <Skeleton className="h-4 w-[200px]" />
              </TableCell>
              <TableCell className="text-right">
                <Skeleton className="h-4 w-12 ml-auto" />
              </TableCell>
              <TableCell className="text-right">
                <Skeleton className="h-4 w-16 ml-auto" />
              </TableCell>
              <TableCell className="text-right">
                <Skeleton className="h-4 w-16 ml-auto" />
              </TableCell>
              <TableCell className="text-right">
                <Skeleton className="h-4 w-16 ml-auto" />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
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
