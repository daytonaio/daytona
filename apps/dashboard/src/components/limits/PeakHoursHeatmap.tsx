import type React from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface HeatmapData {
  day: string
  hour: number
  intensity: number
  usage: number
}

interface PeakHoursHeatmapProps {
  heatmapData: HeatmapData[]
}

export const PeakHoursHeatmap: React.FC<PeakHoursHeatmapProps> = ({ heatmapData }) => {
  const standardDays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  const hours = Array.from({ length: 24 }, (_, i) => i)

  const getIntensityColor = (intensity: number) => {
    if (intensity < 20) return 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600'
    if (intensity < 40) return 'bg-green-300 dark:bg-green-800 hover:bg-green-400 dark:hover:bg-green-700'
    if (intensity < 60) return 'bg-yellow-300 dark:bg-yellow-800 hover:bg-yellow-400 dark:hover:bg-yellow-700'
    if (intensity < 80) return 'bg-orange-300 dark:bg-orange-800 hover:bg-orange-400 dark:hover:bg-orange-700'
    return 'bg-red-300 dark:bg-red-800 hover:bg-red-400 dark:hover:bg-red-700'
  }

  const getDataPoint = (day: string, hour: number) => {
    return heatmapData.find((d) => d.day === day && d.hour === hour)
  }

  const isHourlyView = heatmapData.length <= 24
  const uniqueDays = [...new Set(heatmapData.map((d) => d.day))]
  const daysToShow = isHourlyView ? uniqueDays : uniqueDays.length > 7 ? uniqueDays : standardDays
  const isMonthlyView = uniqueDays.length > 7
  const isDailyAveragesView = isMonthlyView && heatmapData.every((d) => d.hour === 0)

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2">
          <div className="w-5 h-5 rounded bg-muted flex items-center justify-center">
            <div className="w-2 h-2 bg-foreground rounded-full" />
          </div>
          <span>Peak Hours Analysis</span>
        </CardTitle>
        <CardDescription className="text-sm">
          {isHourlyView
            ? 'Hourly usage pattern for the selected time period'
            : isDailyAveragesView
              ? 'Daily usage averages over the past 30 days to identify peak development periods'
              : 'Visual representation of usage patterns throughout the week to identify peak development hours'}
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-0">
        {isHourlyView ? (
          <div className="space-y-3">
            <div className="flex items-center gap-1 mb-2">
              <div className="w-16 text-xs text-muted-foreground font-medium">Hours</div>
              {hours.map((hour) => (
                <div key={hour} className="w-4 text-[10px] text-center text-muted-foreground">
                  {hour % 4 === 0 ? hour.toString().padStart(2, '0') : ''}
                </div>
              ))}
            </div>
            <div className="flex items-center gap-1">
              <div className="w-16 text-xs text-muted-foreground font-medium">{uniqueDays[0]}</div>
              {hours.map((hour) => {
                const dataPoint = getDataPoint(uniqueDays[0], hour)
                const intensity = dataPoint?.intensity || 0
                return (
                  <div
                    key={hour}
                    className={`w-4 h-4 rounded-sm transition-all duration-200 hover:scale-110 cursor-pointer ${getIntensityColor(intensity)}`}
                    title={`${hour}:00 - ${Math.round(intensity)}% usage`}
                  />
                )
              })}
            </div>
          </div>
        ) : isDailyAveragesView ? (
          <div className="space-y-3">
            <div className="text-xs text-muted-foreground font-medium mb-3">Last 30 Days Usage</div>
            <div className="flex flex-wrap gap-1 w-full">
              {uniqueDays.slice(-30).map((day, index) => {
                const dataPoint = getDataPoint(day, 0)
                const intensity = dataPoint?.intensity || 0
                const dayNumber = day.split(' ')[1] || (index + 1).toString()
                return (
                  <div
                    key={day}
                    className={`w-8 h-8 flex items-center justify-center text-xs rounded-sm transition-colors duration-200 cursor-pointer ${getIntensityColor(intensity)}`}
                    title={`${day} - ${Math.round(intensity)}% usage`}
                  >
                    <span className="text-[10px] font-medium text-foreground">{dayNumber}</span>
                  </div>
                )
              })}
            </div>
          </div>
        ) : (
          <div className="space-y-2">
            <div className="flex items-center gap-1 mb-3">
              <div className={isMonthlyView ? 'w-16' : 'w-12'}></div>
              {hours.map((hour) => (
                <div key={hour} className="w-4 text-[10px] text-center text-muted-foreground">
                  {hour % 4 === 0 ? hour.toString().padStart(2, '0') : ''}
                </div>
              ))}
            </div>
            {daysToShow.map((day) => (
              <div key={day} className="flex items-center gap-1">
                <div
                  className={`${isMonthlyView ? 'w-16' : 'w-12'} text-sm text-muted-foreground font-medium truncate`}
                >
                  {day}
                </div>
                {hours.map((hour) => {
                  const dataPoint = getDataPoint(day, hour)
                  const intensity = dataPoint?.intensity || 0
                  return (
                    <div
                      key={`${day}-${hour}`}
                      className={`w-4 h-4 rounded-sm transition-all duration-200 hover:scale-110 cursor-pointer ${getIntensityColor(intensity)}`}
                      title={`${day} ${hour}:00 - ${Math.round(intensity)}% usage`}
                    />
                  )
                })}
              </div>
            ))}
          </div>
        )}

        {/* Legend */}
        <div className="flex items-center justify-between mt-4 text-xs text-muted-foreground">
          <span>Low Usage</span>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 bg-gray-200 dark:bg-gray-700 rounded-sm" />
            <div className="w-3 h-3 bg-green-300 dark:bg-green-800 rounded-sm" />
            <div className="w-3 h-3 bg-yellow-300 dark:bg-yellow-800 rounded-sm" />
            <div className="w-3 h-3 bg-orange-300 dark:bg-orange-800 rounded-sm" />
            <div className="w-3 h-3 bg-red-300 dark:bg-red-800 rounded-sm" />
          </div>
          <span>High Usage</span>
        </div>
      </CardContent>
    </Card>
  )
}

export const PeakHoursHeatmapContent: React.FC<PeakHoursHeatmapProps> = ({ heatmapData }) => {
  const standardDays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  const hours = Array.from({ length: 24 }, (_, i) => i)

  const getIntensityColor = (intensity: number) => {
    if (intensity < 20) return 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600'
    if (intensity < 40) return 'bg-green-300 dark:bg-green-800 hover:bg-green-400 dark:hover:bg-green-700'
    if (intensity < 60) return 'bg-yellow-300 dark:bg-yellow-800 hover:bg-yellow-400 dark:hover:bg-yellow-700'
    if (intensity < 80) return 'bg-orange-300 dark:bg-orange-800 hover:bg-orange-400 dark:hover:bg-orange-700'
    return 'bg-red-300 dark:bg-red-800 hover:bg-red-400 dark:hover:bg-red-700'
  }

  const getDataPoint = (day: string, hour: number) => {
    return heatmapData.find((d) => d.day === day && d.hour === hour)
  }

  const isHourlyView = heatmapData.length <= 24
  const uniqueDays = [...new Set(heatmapData.map((d) => d.day))]
  const daysToShow = isHourlyView ? uniqueDays : uniqueDays.length > 7 ? uniqueDays : standardDays
  const isMonthlyView = uniqueDays.length > 7
  const isDailyAveragesView = isMonthlyView && heatmapData.every((d) => d.hour === 0)

  return (
    <div className="space-y-2">
      <div className="mb-3">
        <h4 className="text-sm font-medium mb-1">Peak Hours Analysis</h4>
        <p className="text-xs text-muted-foreground">
          {isHourlyView
            ? 'Hourly usage pattern for the selected time period'
            : isDailyAveragesView
              ? 'Daily usage averages over the past 30 days to identify peak development periods'
              : 'Visual representation of usage patterns throughout the week to identify peak development hours'}
        </p>
      </div>

      {isHourlyView ? (
        <div className="space-y-3">
          <div className="flex items-center gap-1 mb-2">
            <div className="w-16 text-xs text-muted-foreground font-medium">Hours</div>
            {hours.map((hour) => (
              <div key={hour} className="w-4 text-[10px] text-center text-muted-foreground">
                {hour % 4 === 0 ? hour.toString().padStart(2, '0') : ''}
              </div>
            ))}
          </div>
          <div className="flex items-center gap-1">
            <div className="w-16 text-xs text-muted-foreground font-medium">{uniqueDays[0]}</div>
            {hours.map((hour) => {
              const dataPoint = getDataPoint(uniqueDays[0], hour)
              const intensity = dataPoint?.intensity || 0
              return (
                <div
                  key={hour}
                  className={`w-4 h-4 rounded-sm transition-all duration-200 hover:scale-110 cursor-pointer ${getIntensityColor(intensity)}`}
                  title={`${hour}:00 - ${Math.round(intensity)}% usage`}
                />
              )
            })}
          </div>
        </div>
      ) : isDailyAveragesView ? (
        <div className="space-y-3">
          <div className="text-xs text-muted-foreground font-medium mb-3">Last 30 Days Usage</div>
          <div className="flex flex-wrap gap-1 w-full">
            {uniqueDays.slice(-30).map((day, index) => {
              const dataPoint = getDataPoint(day, 0)
              const intensity = dataPoint?.intensity || 0
              const dayNumber = day.split(' ')[1] || (index + 1).toString()
              return (
                <div
                  key={day}
                  className={`w-8 h-8 flex items-center justify-center text-xs rounded-sm transition-colors duration-200 cursor-pointer ${getIntensityColor(intensity)}`}
                  title={`${day} - ${Math.round(intensity)}% usage`}
                >
                  <span className="text-[10px] font-medium text-foreground">{dayNumber}</span>
                </div>
              )
            })}
          </div>
        </div>
      ) : (
        <div className="space-y-2">
          <div className="flex items-center gap-1 mb-3">
            <div className={isMonthlyView ? 'w-16' : 'w-12'}></div>
            {hours.map((hour) => (
              <div key={hour} className="w-4 text-[10px] text-center text-muted-foreground">
                {hour % 4 === 0 ? hour.toString().padStart(2, '0') : ''}
              </div>
            ))}
          </div>
          {daysToShow.map((day) => (
            <div key={day} className="flex items-center gap-1">
              <div className={`${isMonthlyView ? 'w-16' : 'w-12'} text-sm text-muted-foreground font-medium truncate`}>
                {day}
              </div>
              {hours.map((hour) => {
                const dataPoint = getDataPoint(day, hour)
                const intensity = dataPoint?.intensity || 0
                return (
                  <div
                    key={`${day}-${hour}`}
                    className={`w-4 h-4 rounded-sm transition-all duration-200 hover:scale-110 cursor-pointer ${getIntensityColor(intensity)}`}
                    title={`${day} ${hour}:00 - ${Math.round(intensity)}% usage`}
                  />
                )
              })}
            </div>
          ))}
        </div>
      )}

      {/* Legend */}
      <div className="flex items-center justify-between mt-4 text-xs text-muted-foreground">
        <span>Low Usage</span>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 bg-gray-200 dark:bg-gray-700 rounded-sm" />
          <div className="w-3 h-3 bg-green-300 dark:bg-green-800 rounded-sm" />
          <div className="w-3 h-3 bg-yellow-300 dark:bg-yellow-800 rounded-sm" />
          <div className="w-3 h-3 bg-orange-300 dark:bg-orange-800 rounded-sm" />
          <div className="w-3 h-3 bg-red-300 dark:bg-red-800 rounded-sm" />
        </div>
        <span>High Usage</span>
      </div>
    </div>
  )
}
