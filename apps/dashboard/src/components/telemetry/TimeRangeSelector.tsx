/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useEffect } from 'react'
import { DateRangePicker, DateRangePickerRef, QuickRangesConfig } from '@/components/ui/date-range-picker'
import { DateRange } from 'react-day-picker'
import { subHours } from 'date-fns'

interface TimeRangeSelectorProps {
  onChange: (from: Date, to: Date) => void
  defaultRange?: { from: Date; to: Date }
  defaultSelectedQuickRange?: string
  className?: string
}

const quickRanges: QuickRangesConfig = {
  minutes: [15, 30],
  hours: [1, 3, 6, 12, 24],
  days: [3, 7],
}

export const TimeRangeSelector: React.FC<TimeRangeSelectorProps> = ({
  onChange,
  defaultRange,
  defaultSelectedQuickRange = 'Last 1 hour',
  className,
}) => {
  const pickerRef = React.useRef<DateRangePickerRef>(null)
  const [dateRange, setDateRange] = useState<DateRange>(() => {
    if (defaultRange) {
      return { from: defaultRange.from, to: defaultRange.to }
    }
    // Default to last 1 hour
    const now = new Date()
    return { from: subHours(now, 1), to: now }
  })

  useEffect(() => {
    if (dateRange.from && dateRange.to) {
      onChange(dateRange.from, dateRange.to)
    }
  }, [dateRange, onChange])

  const handleChange = (range: DateRange) => {
    setDateRange(range)
  }

  return (
    <DateRangePicker
      ref={pickerRef}
      value={dateRange}
      onChange={handleChange}
      quickRangesEnabled
      quickRanges={quickRanges}
      timeSelection
      className={className}
      defaultSelectedQuickRange={defaultSelectedQuickRange}
    />
  )
}
