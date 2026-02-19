/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BillableMetricCode, OrganizationUsage } from '@/billing-api/types/OrganizationUsage'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { UsageChart, UsageChartData } from '@/components/UsageChart'
import { AggregatedUsageChart } from '@/components/spending/AggregatedUsageChart'
import { SandboxUsageTable } from '@/components/spending/SandboxUsageTable'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useConfig } from '@/hooks/useConfig'
import { useAggregatedUsage, useSandboxesUsage } from '@/hooks/useAnalyticsUsage'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import { FeatureFlags } from '@/enums/FeatureFlags'
import { useCallback, useEffect, useState } from 'react'
import { DateRangePicker, QuickRangesConfig } from '@/components/ui/date-range-picker'
import { DateRange } from 'react-day-picker'
import { subDays } from 'date-fns'

const analyticsQuickRanges: QuickRangesConfig = {
  hours: [1, 6, 12, 24],
  days: [7, 14, 30],
}

const Spending = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const { billingApi } = useApi()
  const config = useConfig()
  const spendingEnabled = useFeatureFlagEnabled(FeatureFlags.SANDBOX_SPENDING)
  const analyticsAvailable = spendingEnabled && !!config.analyticsApiUrl
  const [currentOrganizationUsage, setCurrentOrganizationUsage] = useState<OrganizationUsage | null>(null)
  const [currentOrganizationUsageLoading, setCurrentOrganizationUsageLoading] = useState(true)
  const [pastOrganizationUsage, setPastOrganizationUsage] = useState<OrganizationUsage[]>([])
  const [pastOrganizationUsageLoading, setPastOrganizationUsageLoading] = useState(true)

  const [analyticsDateRange, setAnalyticsDateRange] = useState<DateRange>(() => {
    const now = new Date()
    return { from: subDays(now, 30), to: now }
  })

  const analyticsParams = {
    from: analyticsDateRange.from ?? subDays(new Date(), 30),
    to: analyticsDateRange.to ?? new Date(),
  }

  const { data: aggregatedUsage, isLoading: aggregatedLoading } = useAggregatedUsage(analyticsParams)
  const { data: sandboxesUsage, isLoading: sandboxesLoading } = useSandboxesUsage(analyticsParams)

  const fetchOrganizationUsage = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setCurrentOrganizationUsageLoading(true)
    try {
      const data = await billingApi.getOrganizationUsage(selectedOrganization.id)
      setCurrentOrganizationUsage(data)
    } catch (error) {
      console.error('Failed to fetch organization usage data:', error)
    } finally {
      setCurrentOrganizationUsageLoading(false)
    }
  }, [billingApi, selectedOrganization])

  const fetchPastOrganizationUsage = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setPastOrganizationUsageLoading(true)
    try {
      const data = await billingApi.getPastOrganizationUsage(selectedOrganization.id)
      setPastOrganizationUsage(data.sort((a, b) => new Date(a.from).getTime() - new Date(b.from).getTime()))
    } catch (error) {
      console.error('Failed to fetch past organization usage data:', error)
    } finally {
      setPastOrganizationUsageLoading(false)
    }
  }, [billingApi, selectedOrganization])

  useEffect(() => {
    if (!selectedOrganization) {
      return
    }
    fetchOrganizationUsage()
    fetchPastOrganizationUsage()
  }, [fetchOrganizationUsage, fetchPastOrganizationUsage, selectedOrganization])

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Spending</PageTitle>
      </PageHeader>

      <PageContent size="full">
        {analyticsAvailable && (
          <div className="space-y-6">
            <div className="flex items-center gap-4">
              <h2 className="text-lg font-semibold shrink-0">Resource Usage</h2>
              <DateRangePicker
                value={analyticsDateRange}
                onChange={setAnalyticsDateRange}
                quickRangesEnabled
                quickRanges={analyticsQuickRanges}
                timeSelection
                defaultSelectedQuickRange="Last 30 days"
                className="w-auto"
              />
            </div>

            <AggregatedUsageChart data={aggregatedUsage} isLoading={aggregatedLoading} />
            <SandboxUsageTable data={sandboxesUsage} isLoading={sandboxesLoading} />
          </div>
        )}

        <UsageChart
          title="Monthly Breakdown"
          usageData={[...pastOrganizationUsage, ...(currentOrganizationUsage ? [currentOrganizationUsage] : [])].map(
            convertUsageToChartData,
          )}
          showTotal
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
