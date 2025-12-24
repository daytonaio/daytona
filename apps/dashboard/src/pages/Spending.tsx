/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BillableMetricCode, OrganizationUsage } from '@/billing-api/types/OrganizationUsage'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { UsageChart, UsageChartData } from '@/components/UsageChart'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useCallback, useEffect, useState } from 'react'

const Spending = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const { billingApi } = useApi()
  const [currentOrganizationUsage, setCurrentOrganizationUsage] = useState<OrganizationUsage | null>(null)
  const [currentOrganizationUsageLoading, setCurrentOrganizationUsageLoading] = useState(true)
  const [pastOrganizationUsage, setPastOrganizationUsage] = useState<OrganizationUsage[]>([])
  const [pastOrganizationUsageLoading, setPastOrganizationUsageLoading] = useState(true)

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
        <UsageChart
          title="Cost Breakdown"
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
