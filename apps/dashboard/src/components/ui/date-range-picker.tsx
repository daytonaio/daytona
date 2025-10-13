/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CalendarIcon } from 'lucide-react'
import { format } from 'date-fns'
import { Calendar } from '@/components/ui/calendar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { DateRange } from 'react-day-picker'
import { useState, forwardRef, useImperativeHandle, useEffect } from 'react'
import { subMinutes, subHours, subDays } from 'date-fns'

// Simple configuration object
export interface QuickRangesConfig {
  minutes?: number[]
  hours?: number[]
  days?: number[]
  months?: number[]
  years?: number[]
}

const createTimeRangesFromConfig = (config: QuickRangesConfig) => {
  const ranges: Array<{ label: string; getRange: () => DateRange }> = []

  // Generate ranges from config
  for (const unit in config) {
    const values = config[unit as keyof QuickRangesConfig]
    if (!values || !Array.isArray(values)) continue

    values.forEach((value) => {
      const unitLabel = value === 1 ? unit.slice(0, -1) : unit // Remove 's' for singular
      ranges.push({
        label: `Last ${value} ${unitLabel}`,
        getRange: () => {
          const now = new Date()
          switch (unit) {
            case 'minutes':
              return { from: subMinutes(now, value), to: now }
            case 'hours':
              return { from: subHours(now, value), to: now }
            case 'days':
              return { from: subDays(now, value), to: now }
            case 'months':
              return { from: subDays(now, value * 30), to: now }
            case 'years':
              return { from: subDays(now, value * 365), to: now }
            default:
              return { from: now, to: now }
          }
        },
      })
    })
  }

  return ranges
}

export interface DateRangePickerProps {
  value?: DateRange
  onChange?: (range: DateRange) => void
  quickRangesEnabled?: boolean
  quickRanges?: QuickRangesConfig
  className?: string
  timeSelection?: boolean
  disabled?: boolean
  defaultSelectedQuickRange?: string
}

export interface DateRangePickerRef {
  getCurrentRange: () => DateRange
}

export const DateRangePicker = forwardRef<DateRangePickerRef, DateRangePickerProps>(
  (
    {
      value,
      onChange,
      quickRangesEnabled = false,
      quickRanges = {},
      className,
      timeSelection = true,
      disabled = false,
      defaultSelectedQuickRange,
    },
    ref,
  ) => {
    const [isOpen, setIsOpen] = useState(false)
    const [selectedQuickRange, setSelectedQuickRange] = useState<string | null>(defaultSelectedQuickRange ?? null)
    const [fromTime, setFromTime] = useState<string>('00:00:00')
    const [toTime, setToTime] = useState<string>('23:59:59')

    // Internal state to track the current selection for preview
    const [internalRange, setInternalRange] = useState<DateRange>(value || { from: undefined, to: undefined })

    // Expose methods to parent component
    useImperativeHandle(
      ref,
      () => ({
        getCurrentRange: (): DateRange => {
          // Always return fresh range for relative selections
          if (selectedQuickRange && selectedQuickRange !== 'All time') {
            const matchingRange = createTimeRangesFromConfig(quickRanges).find(
              (timeRange) => timeRange.label === selectedQuickRange,
            )
            if (matchingRange) {
              return matchingRange.getRange() // Fresh timestamps every time!
            }
          }
          // For custom ranges or "All time", return the current value
          return value || { from: undefined, to: undefined }
        },
      }),
      [selectedQuickRange, quickRanges, value],
    )

    // Sync internal state when parent value changes
    useEffect(() => {
      setInternalRange(value || { from: undefined, to: undefined })
    }, [value])

    const handleQuickRangeSelect = (range: DateRange, label: string) => {
      setSelectedQuickRange(label)
      setInternalRange(range) // Update internal state for preview
      onChange?.(range) // Notify parent immediately
      setIsOpen(false)
    }

    const handleCustomRangeChange = (range: DateRange | undefined) => {
      if (range) {
        setSelectedQuickRange(null) // Clear quick range when using custom
        setInternalRange(range) // Update internal state for preview
        // Removed onChange call - only apply when button is clicked
      }
    }

    const handleTimeChange = (time: string, isFrom: boolean) => {
      if (isFrom) {
        setFromTime(time)
      } else {
        setToTime(time)
      }

      // Removed immediate application - only apply when button is clicked
      // Time changes are now stored in state but not sent to parent until Apply is clicked
    }

    const formatRange = (range: DateRange) => {
      // If a quick range is selected, show its label
      if (selectedQuickRange) return selectedQuickRange

      // Check if this is "All time" (no date restriction)
      if (!range.from && !range.to) return 'Select date range'

      // Helper function to format a date with or without time
      const formatDate = (date: Date) => {
        if (timeSelection) {
          return `${format(date, 'MMM dd, yyyy')}, ${fromTime}`
        }
        return format(date, 'MMM dd, yyyy')
      }

      // Show custom date range with or without time
      if (range.from && range.to) {
        if (timeSelection) {
          // For custom ranges with time, show the actual time that will be applied
          return `${format(range.from, 'MMM dd, yyyy')}, ${fromTime} - ${format(range.to, 'MMM dd, yyyy')}, ${toTime}`
        }
        return `${format(range.from, 'MMM dd, yyyy')} - ${format(range.to, 'MMM dd, yyyy')}`
      } else if (range.from) {
        return `${formatDate(range.from)} - ...`
      }

      return 'Select date range'
    }

    return (
      <Popover open={isOpen} onOpenChange={setIsOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            className={cn(
              'flex w-full justify-between text-left font-normal hover:bg-background',
              !internalRange?.from && 'text-muted-foreground',
              className,
              disabled && 'opacity-50 cursor-not-allowed',
            )}
            disabled={disabled}
          >
            <div className="flex items-center gap-2">
              <CalendarIcon className="h-4 w-4" />
              {internalRange?.from ? formatRange(internalRange) : <span>Select date range</span>}
            </div>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <div className="flex">
            {/* Quick ranges panel */}
            {quickRangesEnabled && (
              <div className="w-64 p-4 border-r">
                <div className="text-sm font-medium mb-3 text-center">Quick ranges</div>
                <div className="space-y-1 max-h-[400px] overflow-y-auto overflow-x-hidden [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-muted-foreground/20 [&::-webkit-scrollbar-thumb]:rounded-full">
                  <button
                    className={cn(
                      'w-full text-left px-3 py-2 text-sm rounded-md transition-colors',
                      selectedQuickRange === 'All time' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted',
                    )}
                    onClick={() => handleQuickRangeSelect({ from: undefined, to: undefined }, 'All time')}
                  >
                    All time
                  </button>
                  {createTimeRangesFromConfig(quickRanges || {}).map((timeRange) => (
                    <button
                      key={timeRange.label}
                      className={cn(
                        'w-full text-left px-3 py-2 text-sm rounded-md transition-colors',
                        selectedQuickRange === timeRange.label
                          ? 'bg-primary text-primary-foreground'
                          : 'hover:bg-muted',
                      )}
                      onClick={() => handleQuickRangeSelect(timeRange.getRange(), timeRange.label)}
                    >
                      {timeRange.label}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Custom range panel */}
            <div className="w-auto p-4">
              <h3 className="font-semibold text-sm mb-3 text-center">Custom range</h3>
              {/* <p className="text-xs text-muted-foreground mb-4">
              Click and drag to select a date range, or click two dates
            </p> */}

              <div className="w-fit mx-auto">
                <Calendar
                  mode="range"
                  selected={internalRange}
                  onSelect={handleCustomRangeChange}
                  numberOfMonths={1}
                  disabled={{ after: new Date() }}
                />
              </div>

              {/* Time selection */}
              {timeSelection && (
                <div className="mt-3">
                  <div className="space-y-3">
                    <div className="flex items-center gap-3">
                      <label className="w-8 text-sm font-medium text-foreground text-left">From</label>
                      <Input
                        type="time"
                        value={fromTime}
                        onChange={(e) => handleTimeChange(e.target.value, true)}
                        step="1"
                        className="w-2/3"
                      />
                    </div>
                    <div className="flex items-center gap-3">
                      <label className="w-8 text-sm font-medium text-foreground text-left">To</label>
                      <Input
                        type="time"
                        value={toTime}
                        onChange={(e) => handleTimeChange(e.target.value, false)}
                        step="1"
                        className="w-2/3"
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Action buttons */}
              <div className="flex justify-center mt-4">
                <Button
                  variant="default"
                  size="sm"
                  className="w-auto px-4"
                  onClick={() => {
                    const applyTimeToDate = (date: Date, timeStr: string) => {
                      const [hours, minutes, seconds] = timeStr.split(':').map((str) => parseInt(str, 10))
                      return new Date(date.getFullYear(), date.getMonth(), date.getDate(), hours, minutes, seconds || 0)
                    }
                    const finalRange: DateRange = {
                      from: applyTimeToDate(internalRange.from || new Date(), fromTime),
                      to: applyTimeToDate(internalRange.to || new Date(), toTime),
                    }
                    onChange?.(finalRange)
                    setIsOpen(false)
                  }}
                >
                  Apply time range
                </Button>
              </div>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    )
  },
)
