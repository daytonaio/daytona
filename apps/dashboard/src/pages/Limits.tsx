'use client'

import type React from 'react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { AlertTriangle, Cpu, HardDrive, MemoryStick, CheckCircle, AlertCircle } from 'lucide-react'
import type { UsageOverview } from '@daytonaio/api-client'
import { handleApiError } from '@/lib/error-handling'
import { Skeleton } from '@/components/ui/skeleton'
import type { OrganizationTier, Tier } from '@/billing-api'
import type { UserProfileIdentity } from './LinkedAccounts'
import { toast } from 'sonner'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import type { OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import { TierTable } from '@/components/TierTable'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tooltip } from '@/components/Tooltip'
import { useApi } from '@/hooks/useApi'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, ResponsiveContainer, ReferenceLine } from 'recharts'
import { PeakHoursHeatmapContent } from '@/components/limits/PeakHoursHeatmap'
import { DateRangePicker } from '@/components/DateRangePicker'

interface HistoricalUsageData {
  timestamp: string
  compute: number
  memory: number
  storage: number
  computePercent: number
  memoryPercent: number
  storagePercent: number
}

type TimePeriod = 'hours' | 'day' | 'week' | 'month' | 'custom'

const Limits: React.FC = () => {
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const { billingApi, organizationsApi } = useApi()
  const [organizationTier, setOrganizationTier] = useState<OrganizationTier | null>(null)
  const [tiers, setTiers] = useState<Tier[]>([])
  const [wallet, setWallet] = useState<OrganizationWallet | null>(null)
  const [usageOverview, setUsage] = useState<UsageOverview | null>(null)
  const [tierLoading, setTierLoading] = useState(false)
  const [historicalUsage, setHistoricalUsage] = useState<HistoricalUsageData[]>([])
  const [selectedTimePeriod, setSelectedTimePeriod] = useState<TimePeriod>('day')
  const [customDateRange, setCustomDateRange] = useState<{ start: string; end: string }>({
    start: '',
    end: '',
  })
  const [chartViewModes, setChartViewModes] = useState<{
    compute: 'chart' | 'heatmap'
    memory: 'chart' | 'heatmap'
    storage: 'chart' | 'heatmap'
  }>({
    compute: 'chart',
    memory: 'chart',
    storage: 'chart',
  })
  const [showDatePicker, setShowDatePicker] = useState(false)

  const mockUsageScenarios = useMemo(
    () => ({
      healthy: {
        currentCpuUsage: 2,
        totalCpuQuota: 10,
        currentMemoryUsage: 3,
        totalMemoryQuota: 10,
        currentDiskUsage: 8,
        totalDiskQuota: 30,
        totalGpuQuota: 0,
      },
      warning: {
        currentCpuUsage: 7,
        totalCpuQuota: 10,
        currentMemoryUsage: 8,
        totalMemoryQuota: 10,
        currentDiskUsage: 25,
        totalDiskQuota: 30,
        totalGpuQuota: 0,
      },
      critical: {
        currentCpuUsage: 9,
        totalCpuQuota: 10,
        currentMemoryUsage: 9.5,
        totalMemoryQuota: 10,
        currentDiskUsage: 28,
        totalDiskQuota: 30,
        totalGpuQuota: 0,
      },
      mixed: {
        currentCpuUsage: 1,
        totalCpuQuota: 10,
        currentMemoryUsage: 6,
        totalMemoryQuota: 10,
        currentDiskUsage: 28,
        totalDiskQuota: 30,
        totalGpuQuota: 0,
      },
    }),
    [],
  )

  const mockOrganizationTier = useMemo(
    () => ({
      tier: 2,
      id: 'mock-tier-id',
      largestSuccessfulPaymentCents: 5000,
      hasVerifiedBusinessEmail: true,
    }),
    [],
  )

  const mockTiers = useMemo(
    () => [
      {
        tier: 1,
        name: 'Starter',
        tierLimit: {
          concurrentCPU: 5,
          concurrentRAMGiB: 5,
          concurrentDiskGiB: 15,
        },
        minTopUpAmountCents: 0,
        topUpIntervalDays: 0,
      },
      {
        tier: 2,
        name: 'Pro',
        tierLimit: {
          concurrentCPU: 10,
          concurrentRAMGiB: 10,
          concurrentDiskGiB: 30,
        },
        minTopUpAmountCents: 2500, // $25
        topUpIntervalDays: 0,
      },
      {
        tier: 3,
        name: 'Enterprise',
        tierLimit: {
          concurrentCPU: 20,
          concurrentRAMGiB: 20,
          concurrentDiskGiB: 60,
        },
        minTopUpAmountCents: 10000, // $100
        topUpIntervalDays: 30,
      },
    ],
    [],
  )

  const mockWallet = useMemo(
    () => ({
      creditCardConnected: true,
      balance: 25.0,
      balanceCents: 2500,
      ongoingBalanceCents: 0,
      name: 'Mock Organization Wallet',
    }),
    [],
  )

  const useMockData = true // Set to false to use real API calls
  const mockScenario = 'mixed' // Change to: 'healthy', 'warning', 'critical', 'mixed'

  const fetchOrganizationTier = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setTierLoading(true)
    try {
      if (useMockData) {
        await new Promise((resolve) => setTimeout(resolve, 500))
        setOrganizationTier(mockOrganizationTier)
      } else {
        const data = await billingApi.getOrganizationTier(selectedOrganization.id)
        setOrganizationTier(data)
      }
    } catch (error) {
      handleApiError(error, 'Failed to fetch organization tier')
    } finally {
      setTierLoading(false)
    }
  }, [selectedOrganization, billingApi, mockOrganizationTier, useMockData])

  const fetchTiers = useCallback(async () => {
    try {
      if (useMockData) {
        await new Promise((resolve) => setTimeout(resolve, 300))
        setTiers(mockTiers)
      } else {
        const data = await billingApi.listTiers()
        setTiers(data)
      }
    } catch (error) {
      handleApiError(error, 'Failed to fetch tiers')
    }
  }, [billingApi, mockTiers, useMockData])

  const fetchOrganizationWallet = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    try {
      if (useMockData) {
        await new Promise((resolve) => setTimeout(resolve, 200))
        setWallet(mockWallet)
      } else {
        const data = await billingApi.getOrganizationWallet(selectedOrganization.id)
        setWallet(data)
      }
    } catch (error) {
      handleApiError(error, 'Failed to fetch organization wallet')
    }
  }, [selectedOrganization, billingApi, mockWallet, useMockData])

  const fetchUsage = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    try {
      if (useMockData) {
        await new Promise((resolve) => setTimeout(resolve, 100))
        setUsage(mockUsageScenarios[mockScenario])
      } else {
        const response = await organizationsApi.getOrganizationUsageOverview(selectedOrganization.id)
        setUsage(response.data)
      }
    } catch (error) {
      handleApiError(error, 'Failed to fetch usage data')
    }
  }, [selectedOrganization, organizationsApi, mockUsageScenarios, mockScenario, useMockData])

  const handleUpgrade = useCallback(
    async (tier: number) => {
      if (!selectedOrganization) {
        return
      }

      try {
        await billingApi.upgradeTier(selectedOrganization.id, tier)
        toast.success('Tier upgraded successfully')
        fetchOrganizationTier()
        fetchUsage()
      } catch (error) {
        handleApiError(error, 'Failed to upgrade organization tier')
      }
    },
    [selectedOrganization, billingApi, fetchOrganizationTier, fetchUsage],
  )

  const handleDowngrade = useCallback(
    async (tier: number) => {
      if (!selectedOrganization) {
        return
      }

      try {
        await billingApi.downgradeTier(selectedOrganization.id, tier)
        toast.success('Tier downgraded successfully')
        fetchOrganizationTier()
        fetchUsage()
      } catch (error) {
        handleApiError(error, 'Failed to downgrade organization tier')
      }
    },
    [selectedOrganization, billingApi, fetchOrganizationTier, fetchUsage],
  )

  const generateMockHistoricalData = useCallback(
    (period: TimePeriod): HistoricalUsageData[] => {
      const data: HistoricalUsageData[] = []
      const now = new Date()
      let points: number

      switch (period) {
        case 'hours':
          points = 24 // Rolling 24 hours from current time
          break
        case 'day':
          points = now.getHours() + 1 // From midnight to current hour
          break
        case 'week': {
          points = 7 // Last 7 days
          break
        }
        case 'month':
          points = 30 // Last 30 days
          break
        case 'custom':
          points = 10 // Custom range for demonstration
          break
      }

      const currentTier = mockTiers.find((t) => t.tier === mockOrganizationTier.tier)
      const computeLimit = currentTier?.tierLimit.concurrentCPU || 10
      const memoryLimit = currentTier?.tierLimit.concurrentRAMGiB || 10
      const storageLimit = currentTier?.tierLimit.concurrentDiskGiB || 30

      for (let i = points - 1; i >= 0; i--) {
        let timestamp: string
        switch (period) {
          case 'hours': {
            const hoursAgo = new Date(now.getTime() - i * 60 * 60 * 1000)
            timestamp = `${hoursAgo.getHours().toString().padStart(2, '0')}:00`
            break
          }
          case 'day':
            timestamp = `${(points - 1 - i).toString().padStart(2, '0')}:00`
            break
          case 'week': {
            const weekdays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
            timestamp = weekdays[6 - i] || `Day ${7 - i}`
            break
          }
          case 'month':
            timestamp = (30 - i).toString()
            break
          case 'custom':
            timestamp = (points - i).toString()
            break
          default:
            timestamp = i.toString()
        }

        let baseCompute: number
        let baseMemory: number
        let baseStorage: number

        if (period === 'hours') {
          // Rolling 24h - more realistic continuous pattern
          baseCompute = 1 + Math.sin((24 - i) * 0.3) * 0.5 + Math.random() * 0.3
          baseMemory = 6 + Math.sin((24 - i) * 0.2) * 2 + Math.random() * 0.5
          baseStorage = 25 + Math.sin((24 - i) * 0.1) * 3 + Math.random() * 1
        } else if (period === 'day') {
          // Calendar day - show progression from midnight to now
          const hourOfDay = points - 1 - i
          const workHourMultiplier = hourOfDay >= 9 && hourOfDay <= 17 ? 1.5 : 0.7
          baseCompute = (1 + Math.sin(hourOfDay * 0.3) * 0.5 + Math.random() * 0.3) * workHourMultiplier
          baseMemory = (6 + Math.sin(hourOfDay * 0.2) * 2 + Math.random() * 0.5) * workHourMultiplier
          baseStorage = (25 + Math.sin(hourOfDay * 0.1) * 3 + Math.random() * 1) * workHourMultiplier
        } else {
          // Week/month - existing logic
          baseCompute = 1 + Math.sin(i * 0.3) * 0.5 + Math.random() * 0.3
          baseMemory = 6 + Math.sin(i * 0.2) * 2 + Math.random() * 0.5
          baseStorage = 25 + Math.sin(i * 0.1) * 3 + Math.random() * 1
        }

        const compute = Math.max(0, Math.min(computeLimit, baseCompute))
        const memory = Math.max(0, Math.min(memoryLimit, baseMemory))
        const storage = Math.max(0, Math.min(storageLimit, baseStorage))

        data.push({
          timestamp,
          compute,
          memory,
          storage,
          computePercent: (compute / computeLimit) * 100,
          memoryPercent: (memory / memoryLimit) * 100,
          storagePercent: (storage / storageLimit) * 100,
        })
      }

      return data
    },
    [mockTiers, mockOrganizationTier],
  )

  const fetchHistoricalUsage = useCallback(
    async (period: TimePeriod) => {
      if (!selectedOrganization) {
        return
      }
      try {
        if (useMockData) {
          await new Promise((resolve) => setTimeout(resolve, 300))
          setHistoricalUsage(generateMockHistoricalData(period))
        } else {
          setHistoricalUsage(generateMockHistoricalData(period))
        }
      } catch (error) {
        handleApiError(error, 'Failed to fetch historical usage data')
      }
    },
    [selectedOrganization, generateMockHistoricalData, useMockData],
  )

  const generateResourceHeatmapData = useCallback(
    (resourceType: 'compute' | 'memory' | 'storage', period: TimePeriod) => {
      const heatmapData = []

      if (period === 'hours') {
        const hours = Array.from({ length: 24 }, (_, i) => i)
        for (const hour of hours) {
          const isWorkHour = hour >= 9 && hour <= 17
          const isEveningHour = hour >= 18 && hour <= 22
          let intensity = Math.random() * 15 // Base random usage

          if (resourceType === 'compute') {
            if (isWorkHour) intensity += 45 + Math.random() * 35
            else if (isEveningHour) intensity += 25 + Math.random() * 25
            else intensity += Math.random() * 20
          } else if (resourceType === 'memory') {
            if (isWorkHour) intensity += 40 + Math.random() * 30
            else if (isEveningHour) intensity += 20 + Math.random() * 20
            else intensity += Math.random() * 25
          } else if (resourceType === 'storage') {
            intensity += 30 + hour * 1.5 + Math.random() * 20
          }

          heatmapData.push({
            day: 'Last 24h',
            hour,
            intensity: Math.min(100, intensity),
            usage: Math.min(100, intensity),
          })
        }
      } else if (period === 'day') {
        const currentHour = new Date().getHours()
        const hours = Array.from({ length: 24 }, (_, i) => i)

        for (const hour of hours) {
          const isWorkHour = hour >= 9 && hour <= 17
          const isPastHour = hour <= currentHour
          let intensity = isPastHour ? Math.random() * 20 : 0 // No data for future hours

          if (isPastHour && resourceType === 'compute') {
            if (isWorkHour) intensity += 35 + Math.random() * 30
            else intensity += Math.random() * 15
          } else if (isPastHour && resourceType === 'memory') {
            if (isWorkHour) intensity += 30 + Math.random() * 25
            else intensity += Math.random() * 20
          } else if (isPastHour && resourceType === 'storage') {
            intensity += 25 + hour * 2 + Math.random() * 15
          }

          heatmapData.push({
            day: 'Today',
            hour,
            intensity: Math.min(100, intensity),
            usage: Math.min(100, intensity),
          })
        }
      } else if (period === 'week') {
        const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
        const hours = Array.from({ length: 24 }, (_, i) => i)

        for (const day of days) {
          for (const hour of hours) {
            const isWeekday = !['Sat', 'Sun'].includes(day)
            const isWorkHour = hour >= 9 && hour <= 17
            const dayIndex = days.indexOf(day)
            let intensity = Math.random() * 20

            if (resourceType === 'compute') {
              if (isWeekday && isWorkHour) {
                const midWeekBoost = dayIndex >= 2 && dayIndex <= 3 ? 15 : 0
                intensity += 40 + midWeekBoost + Math.random() * 25
              } else if (isWeekday) {
                intensity += 15 + Math.random() * 15
              } else {
                intensity += Math.random() * 10
              }
            } else if (resourceType === 'memory') {
              if (isWeekday) {
                const tuesdayBoost = dayIndex === 1 ? 10 : 0
                intensity += 25 + tuesdayBoost + Math.random() * 20
                if (isWorkHour) intensity += 15
              } else {
                intensity += Math.random() * 15
              }
            } else if (resourceType === 'storage') {
              intensity += dayIndex * 8 + Math.random() * 15
              if (isWorkHour) intensity += 8
            }

            heatmapData.push({
              day,
              hour,
              intensity: Math.min(100, intensity),
              usage: Math.min(100, intensity),
            })
          }
        }
      } else if (period === 'month') {
        const today = new Date()
        const days = []

        // Create last 30 days
        for (let i = 29; i >= 0; i--) {
          const date = new Date(today)
          date.setDate(date.getDate() - i)
          const dayName = date.toLocaleDateString('en-US', { weekday: 'short' })
          const dayNumber = date.getDate()
          days.push(`${dayName} ${dayNumber}`)
        }

        // Generate one average value per day instead of 24 hourly values
        for (const day of days) {
          const dayIndex = days.indexOf(day)
          const isWeekend = day.startsWith('Sat') || day.startsWith('Sun')
          let dailyAverage = Math.random() * 25 // Base daily usage

          if (resourceType === 'compute') {
            if (!isWeekend) {
              dailyAverage += 45 + Math.random() * 30 // Weekday usage
              // Add weekly patterns
              const weekOfMonth = Math.floor(dayIndex / 7)
              if (weekOfMonth === 1 || weekOfMonth === 2) dailyAverage += 10 // Mid-month peak
            } else {
              dailyAverage += Math.random() * 15 // Weekend usage
            }
          } else if (resourceType === 'memory') {
            if (!isWeekend) {
              dailyAverage += 50 + Math.random() * 25 // High sustained memory usage
              // Memory builds up over the month
              dailyAverage += dayIndex * 0.8
            } else {
              dailyAverage += Math.random() * 20
            }
          } else if (resourceType === 'storage') {
            dailyAverage += 35 + dayIndex * 1.5 + Math.random() * 20 // Storage grows over time
          }

          // Store as single data point per day (hour 0 represents daily average)
          heatmapData.push({
            day,
            hour: 0, // Use hour 0 to represent daily average
            intensity: Math.min(100, dailyAverage),
            usage: Math.min(100, dailyAverage),
          })
        }
      }

      return heatmapData
    },
    [],
  )

  const computeHeatmapData = useMemo(
    () => generateResourceHeatmapData('compute', selectedTimePeriod),
    [generateResourceHeatmapData, selectedTimePeriod],
  )
  const memoryHeatmapData = useMemo(
    () => generateResourceHeatmapData('memory', selectedTimePeriod),
    [generateResourceHeatmapData, selectedTimePeriod],
  )
  const storageHeatmapData = useMemo(
    () => generateResourceHeatmapData('storage', selectedTimePeriod),
    [generateResourceHeatmapData, selectedTimePeriod],
  )

  const githubConnected = useMemo(() => {
    if (!user?.profile?.identities) {
      return false
    }
    return (user.profile.identities as UserProfileIdentity[]).some(
      (identity: UserProfileIdentity) => identity.provider === 'github',
    )
  }, [user])

  useEffect(() => {
    fetchOrganizationTier().finally(() => fetchUsage())
    fetchTiers()
    fetchHistoricalUsage(selectedTimePeriod)
    const interval = setInterval(fetchUsage, 10000)
    return () => clearInterval(interval)
  }, [fetchOrganizationTier, fetchUsage, fetchTiers, fetchHistoricalUsage, selectedTimePeriod])

  useEffect(() => {
    fetchHistoricalUsage(selectedTimePeriod)
  }, [selectedTimePeriod, fetchHistoricalUsage])

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Limits</h1>
      </div>

      <Card className="my-4">
        <CardHeader>
          <CardTitle className="flex items-center mb-2">
            Usage Limits{' '}
            {organizationTier && (
              <Badge variant="outline" className="ml-2 text-sm">
                Tier {organizationTier.tier}
              </Badge>
            )}
          </CardTitle>
          <CardDescription>
            Limits help us mitigate misuse and manage infrastructure resources. Ensuring fair and stable access to
            sandboxes and compute capacity across all users.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {!usageOverview && (
            <div className="flex items-center justify-center h-full">
              <Skeleton className="w-full h-full" />
            </div>
          )}
          {usageOverview && (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
              {['Compute', 'Memory', 'Storage'].map((resourceName) => {
                const resource = {
                  name: resourceName,
                  icon:
                    resourceName === 'Compute' ? (
                      <Cpu className="w-5 h-5" />
                    ) : resourceName === 'Memory' ? (
                      <MemoryStick className="w-5 h-5" />
                    ) : (
                      <HardDrive className="w-5 h-5" />
                    ),
                  description:
                    resourceName === 'Compute'
                      ? 'CPU cores allocated to your workspaces'
                      : resourceName === 'Memory'
                        ? 'RAM allocated to your workspaces'
                        : 'Disk space used by your workspaces',
                  current:
                    resourceName === 'Compute'
                      ? usageOverview.currentCpuUsage || 0
                      : resourceName === 'Memory'
                        ? usageOverview.currentMemoryUsage || 0
                        : usageOverview.currentDiskUsage || 0,
                  total:
                    resourceName === 'Compute'
                      ? usageOverview.totalCpuQuota || 0
                      : resourceName === 'Memory'
                        ? usageOverview.totalMemoryQuota || 0
                        : usageOverview.totalDiskQuota || 0,
                  unit: resourceName === 'Storage' ? 'GiB' : 'vCPU',
                  dataKey: resourceName === 'Compute' ? 'compute' : resourceName === 'Memory' ? 'memory' : 'storage',
                  percentKey:
                    resourceName === 'Compute'
                      ? 'computePercent'
                      : resourceName === 'Memory'
                        ? 'memoryPercent'
                        : 'storagePercent',
                  color:
                    resourceName === 'Compute'
                      ? '#10b981' // green-500
                      : resourceName === 'Memory'
                        ? '#3b82f6' // blue-500
                        : '#ef4444', // red-500
                }

                const percentage = (resource.current / resource.total) * 100
                const isHighUsage = percentage > 90
                const isMediumUsage = percentage >= 70 && percentage <= 90

                const getStatusColor = () => {
                  if (isHighUsage) return 'text-red-500'
                  if (isMediumUsage) return 'text-yellow-500'
                  return 'text-green-500'
                }

                const getStatusIcon = () => {
                  if (isHighUsage) return <AlertTriangle className="w-4 h-4 text-red-500" />
                  if (isMediumUsage) return <AlertCircle className="w-4 h-4 text-yellow-500" />
                  return <CheckCircle className="w-4 h-4 text-green-500" />
                }

                const getStatusText = () => {
                  if (isHighUsage) return 'Critical'
                  if (isMediumUsage) return 'Warning'
                  return 'Healthy'
                }

                return (
                  <Card key={resource.name} className="p-3">
                    <div>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          {resource.icon}
                          <h3 className="font-semibold text-base">{resource.name}</h3>
                        </div>
                        <div className="flex items-center gap-2">
                          {getStatusIcon()}
                          <Badge variant="outline" className={`text-xs ${getStatusColor()}`}>
                            {getStatusText()}
                          </Badge>
                        </div>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">{resource.description}</p>
                      <div className="border-t-2 border-border/60 my-6"></div>
                      <div className="flex items-center justify-between mt-3">
                        <div className="flex-1">
                          <div className="text-lg font-mono">
                            {resource.current}/{resource.total}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {resource.unit} • {Math.round(percentage)}% used
                          </div>
                        </div>
                        <div className="flex flex-col items-center gap-1">
                          <div className="w-24 h-8 flex flex-col gap-0.5">
                            {Array.from({ length: 10 }).map((_, index) => {
                              const barThreshold = ((index + 1) / 10) * 100
                              const isActive = percentage >= barThreshold - 10
                              return (
                                <div
                                  key={index}
                                  className={`h-0.5 w-full rounded-full transition-all duration-300 ${
                                    isActive
                                      ? isHighUsage
                                        ? 'bg-red-500'
                                        : isMediumUsage
                                          ? 'bg-yellow-500'
                                          : 'bg-green-500'
                                      : 'bg-gray-200 dark:bg-gray-700'
                                  }`}
                                />
                              )
                            })}
                          </div>
                        </div>
                      </div>
                    </div>
                  </Card>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="my-4">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-start gap-3 mb-2">
                <div className="flex items-center gap-2">
                  <div className="mt-0.5">
                    <div className="w-5 h-5 rounded bg-muted flex items-center justify-center">
                      <div className="w-2 h-2 bg-foreground rounded-full" />
                    </div>
                  </div>
                  <span>Usage Over Time</span>
                </div>
              </CardTitle>
              <CardDescription>
                Track your resource usage patterns to understand when you might need to upgrade to a higher tier.
              </CardDescription>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex gap-1 border border-border rounded-lg p-0.5 bg-muted/30">
                <Button
                  variant={Object.values(chartViewModes).every((mode) => mode === 'chart') ? 'default' : 'secondary'}
                  size="sm"
                  onClick={() => setChartViewModes({ compute: 'chart', memory: 'chart', storage: 'chart' })}
                >
                  Chart
                </Button>
                <Button
                  variant={Object.values(chartViewModes).every((mode) => mode === 'heatmap') ? 'default' : 'secondary'}
                  size="sm"
                  onClick={() => setChartViewModes({ compute: 'heatmap', memory: 'heatmap', storage: 'heatmap' })}
                >
                  Heatmap
                </Button>
              </div>
              <div className="flex gap-2 relative">
                {['hours', 'day', 'week', 'month', 'custom'].map((period) => {
                  const getTooltipContent = () => {
                    switch (period) {
                      case 'hours':
                        return 'Rolling 24 hours from current time'
                      case 'day':
                        return 'Current calendar day from midnight to now'
                      case 'week':
                        return 'Last 7 days'
                      case 'month':
                        return 'Last 30 days'
                      case 'custom':
                        return 'Select custom date range'
                      default:
                        return ''
                    }
                  }

                  const getButtonLabel = () => {
                    switch (period) {
                      case 'hours':
                        return '24h'
                      case 'week':
                        return '7d'
                      case 'month':
                        return '30d'
                      case 'custom':
                        return customDateRange.start && customDateRange.end
                          ? `${customDateRange.start} - ${customDateRange.end}`
                          : 'Custom'
                      default:
                        return 'Day'
                    }
                  }

                  return (
                    <div key={period} className="relative">
                      <Tooltip
                        label={
                          <Button
                            variant={selectedTimePeriod === period ? 'default' : 'outline'}
                            size="sm"
                            onClick={() => {
                              if (period === 'custom') {
                                setShowDatePicker(!showDatePicker)
                              } else {
                                setSelectedTimePeriod(period as TimePeriod)
                                setShowDatePicker(false)
                              }
                            }}
                          >
                            {getButtonLabel()}
                          </Button>
                        }
                        content={<span className="text-xs">{getTooltipContent()}</span>}
                      />
                      {period === 'custom' && showDatePicker && (
                        <DateRangePicker
                          onRangeSelect={(start, end) => {
                            setCustomDateRange({ start, end })
                            setSelectedTimePeriod('custom')
                            setShowDatePicker(false)
                          }}
                          onClose={() => setShowDatePicker(false)}
                        />
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {historicalUsage.length === 0 ? (
            <div className="flex items-center justify-center h-64">
              <Skeleton className="w-full h-full" />
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {['Compute', 'Memory', 'Storage'].map((resourceName) => {
                const resourceKey = resourceName.toLowerCase() as 'compute' | 'memory' | 'storage'
                const resource = {
                  name: resourceName,
                  icon:
                    resourceName === 'Compute' ? (
                      <Cpu className="w-5 h-5" />
                    ) : resourceName === 'Memory' ? (
                      <MemoryStick className="w-5 h-5" />
                    ) : (
                      <HardDrive className="w-5 h-5" />
                    ),
                  dataKey:
                    resourceName === 'Compute'
                      ? 'computePercent'
                      : resourceName === 'Memory'
                        ? 'memoryPercent'
                        : 'storagePercent',
                  percentKey: resourceName === 'Compute' ? 'compute' : resourceName === 'Memory' ? 'memory' : 'storage',
                  color:
                    resourceName === 'Compute'
                      ? '#10b981' // green-500
                      : resourceName === 'Memory'
                        ? '#3b82f6' // blue-500
                        : '#ef4444', // red-500
                  unit: resourceName === 'Storage' ? 'GiB' : 'vCPU',
                  heatmapData:
                    resourceKey === 'compute'
                      ? computeHeatmapData
                      : resourceKey === 'memory'
                        ? memoryHeatmapData
                        : storageHeatmapData,
                }

                const currentValue =
                  historicalUsage[historicalUsage.length - 1]?.[resource.percentKey as keyof HistoricalUsageData] || 0
                const currentPercent =
                  historicalUsage[historicalUsage.length - 1]?.[resource.dataKey as keyof HistoricalUsageData] || 0

                const isHeatmapView = chartViewModes[resourceKey] === 'heatmap'

                return (
                  <Card key={resource.name} className="p-4">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-start gap-3">
                        <div className="mt-0.5">{resource.icon}</div>
                        <div>
                          <h3 className="font-semibold text-base">{resource.name}</h3>
                          <p className="text-sm text-muted-foreground">
                            {selectedTimePeriod === 'hours' || selectedTimePeriod === 'day'
                              ? '24h'
                              : selectedTimePeriod === 'week'
                                ? '7d'
                                : selectedTimePeriod === 'month'
                                  ? '30d'
                                  : 'Custom'}
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-lg font-mono">
                          {typeof currentPercent === 'number' ? currentPercent.toFixed(0) : currentPercent}%
                        </div>
                        <div className="text-sm text-muted-foreground">
                          {typeof currentValue === 'number' ? currentValue.toFixed(1) : currentValue} {resource.unit}
                        </div>
                      </div>
                    </div>

                    {isHeatmapView ? (
                      <div className="h-auto py-4">
                        <PeakHoursHeatmapContent heatmapData={resource.heatmapData} />
                      </div>
                    ) : (
                      <div className="h-32">
                        <ResponsiveContainer width="100%" height="100%">
                          <LineChart data={historicalUsage}>
                            <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                            <XAxis
                              dataKey="timestamp"
                              axisLine={false}
                              tickLine={false}
                              tick={{ fontSize: 12, opacity: 0.6 }}
                              className="text-muted-foreground/60"
                              interval="preserveStartEnd"
                            />
                            <YAxis
                              axisLine={false}
                              tickLine={false}
                              tick={{ fontSize: 12, opacity: 0.6 }}
                              className="text-muted-foreground/60"
                              domain={[0, 100]}
                              tickFormatter={(value) => `${value}%`}
                            />
                            <ReferenceLine
                              y={100}
                              stroke="#ef4444"
                              strokeDasharray="2 2"
                              strokeOpacity={0.6}
                              label={{ value: 'Tier Limit', fontSize: 10, fill: '#ef4444' }}
                            />
                            <ReferenceLine y={90} stroke="#f59e0b" strokeDasharray="1 1" strokeOpacity={0.4} />
                            <ReferenceLine y={80} stroke="#10b981" strokeDasharray="1 1" strokeOpacity={0.3} />
                            <Line
                              type="monotone"
                              dataKey={resource.dataKey}
                              stroke={resource.color}
                              strokeWidth={2}
                              dot={false}
                              activeDot={{ r: 4, fill: resource.color }}
                            />
                          </LineChart>
                        </ResponsiveContainer>
                        <div className="flex items-center justify-between mt-6 mb-4 text-xs text-muted-foreground">
                          <span>Optimal (&lt;80%)</span>
                          <span>Warning (80-90%)</span>
                          <span>Critical (&gt;90%)</span>
                        </div>
                      </div>
                    )}
                  </Card>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="border-t-2 border-border/60 my-6"></div>

      <Card className="my-4">
        <CardHeader>
          <CardTitle className="flex items-center mb-2">Increasing your limits</CardTitle>
          <CardDescription>
            {organizationTier ? (
              <>
                Your organization is currently in <b>Tier {organizationTier.tier}</b>. Your limits will automatically be
                increased once you move to the next tier based on the criteria outlined below.
                <br />
                Note: For the top up requirements, make sure to top up in a single transaction.
              </>
            ) : (
              'Loading tier information...'
            )}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {tierLoading ? (
            <div className="flex items-center justify-center h-32">
              <Skeleton className="w-full h-full" />
            </div>
          ) : organizationTier ? (
            <TierTable
              emailVerified={!!user?.profile?.email_verified}
              githubConnected={githubConnected}
              organizationTier={organizationTier}
              creditCardConnected={!!wallet?.creditCardConnected}
              phoneVerified={!!user?.profile?.phone_verified}
              tierLoading={tierLoading}
              tiers={tiers}
              onUpgrade={handleUpgrade}
              onDowngrade={handleDowngrade}
            />
          ) : (
            <div className="text-center text-muted-foreground py-8">
              Unable to load tier information. Please try refreshing the page.
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default Limits
