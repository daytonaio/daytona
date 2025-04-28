import React, { useCallback, useEffect, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { UsageOverview } from '@daytonaio/api-client'
import { AlertTriangle } from 'lucide-react'
import QuotaLine from '@/components/QuotaLine'
import { Card, CardContent } from '@/components/ui/card'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'

const Usage: React.FC = () => {
  const { organizationsApi } = useApi()
  const [usageOverview, setUsage] = useState<UsageOverview | null>(null)
  const { selectedOrganization } = useSelectedOrganization()

  const fetchUsage = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    try {
      const response = await organizationsApi.getOrganizationUsageOverview(selectedOrganization.id)
      setUsage(response.data)
    } catch (error) {
      handleApiError(error, 'Failed to fetch usage data')
    }
  }, [organizationsApi, selectedOrganization])

  useEffect(() => {
    fetchUsage()
    const interval = setInterval(fetchUsage, 10000)
    return () => clearInterval(interval)
  }, [fetchUsage])

  const getUsageDisplay = (current: number, total: number, unit = '') => {
    const percentage = (current / total) * 100
    const isHighUsage = percentage > 90

    return (
      <div className="flex items-center gap-1">
        <span className={isHighUsage ? 'text-red-500' : undefined}>
          {current}/{total}
          {unit}
        </span>
        {isHighUsage && <AlertTriangle className="w-4 h-4 text-red-500" />}
      </div>
    )
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold">Usage</h1>
      {usageOverview && (
        <Card className="my-4">
          <CardContent className="p-6">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="flex items-center justify-between mb-1">
                  <span>Sandboxes:</span>
                  {getUsageDisplay(usageOverview.currentWorkspaces, usageOverview.totalWorkspaceQuota)}
                </div>
                <QuotaLine current={usageOverview.currentWorkspaces} total={usageOverview.totalWorkspaceQuota} />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <span>Running Sandboxes:</span>
                  {getUsageDisplay(usageOverview.concurrentWorkspaces, usageOverview.concurrentWorkspaceQuota)}
                </div>
                <QuotaLine
                  current={usageOverview.concurrentWorkspaces}
                  total={usageOverview.concurrentWorkspaceQuota}
                />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1 mt-3">
                  <span>CPU:</span>
                  {getUsageDisplay(usageOverview.currentCpuUsage, usageOverview.totalCpuQuota)}
                </div>
                <QuotaLine current={usageOverview.currentCpuUsage} total={usageOverview.totalCpuQuota} />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1 mt-3">
                  <span>Memory:</span>
                  {getUsageDisplay(usageOverview.currentMemoryUsage, usageOverview.totalMemoryQuota, 'GB')}
                </div>
                <QuotaLine current={usageOverview.currentMemoryUsage} total={usageOverview.totalMemoryQuota} />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1 mt-3">
                  <span>Disk:</span>
                  {getUsageDisplay(usageOverview.currentDiskUsage, usageOverview.totalDiskQuota, 'GB')}
                </div>
                <QuotaLine current={usageOverview.currentDiskUsage} total={usageOverview.totalDiskQuota} />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1 mt-3">
                  <span>Images:</span>
                  {getUsageDisplay(usageOverview.currentImageNumber, usageOverview.imageQuota)}
                </div>
                <QuotaLine current={usageOverview.currentImageNumber} total={usageOverview.imageQuota} />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1 mt-3">
                  <span>Total Images Size:</span>
                  {getUsageDisplay(
                    Number(usageOverview.totalImageSizeUsed.toFixed(1)),
                    usageOverview.totalImageSizeQuota,
                    'GB',
                  )}
                </div>
                <QuotaLine current={usageOverview.totalImageSizeUsed} total={usageOverview.totalImageSizeQuota} />
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

export default Usage
