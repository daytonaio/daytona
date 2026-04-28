/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useMemo, useState, useEffect } from 'react'
import { DateRangePicker, DateRangePickerRef, QuickRangesConfig } from '@/components/ui/date-range-picker'
import { DateRange } from 'react-day-picker'
import { subDays, subHours, subMinutes } from 'date-fns'

interface TimeRangeSelectorProps {
  onChange: (from: Date, to: Date) => void
  onClear?: () => void
  defaultRange?: { from: Date; to: Date }
  defaultSelectedQuickRange?: string
  className?: string
}

const quickRanges: QuickRangesConfig = {
  minutes: [15, 30],
  hours: [1, 3, 6, 12, 24],
  days: [3, 7],
}

function getQuickRange(label: string | undefined, config: QuickRangesConfig): DateRange | null {
  if (!label || label === 'All time') {
    return null
  }

  for (const unit in config) {
    const values = config[unit as keyof QuickRangesConfig]
    if (!values || !Array.isArray(values)) continue

    for (const value of values) {
      const unitLabel = value === 1 ? unit.slice(0, -1) : unit
      if (label !== `Last ${value} ${unitLabel}`) {
        continue
      }

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
      }
    }
  }

  return null
}

function getQuickRangeLabel(range: DateRange, config: QuickRangesConfig, now = new Date()) {
  if (!range.from || !range.to) {
    return undefined
  }

  const toDeltaMs = Math.abs(range.to.getTime() - now.getTime())

  if (toDeltaMs > 60_000) {
    return undefined
  }

  for (const unit in config) {
    const values = config[unit as keyof QuickRangesConfig]
    if (!values || !Array.isArray(values)) continue

    for (const value of values) {
      const unitLabel = value === 1 ? unit.slice(0, -1) : unit
      const label = `Last ${value} ${unitLabel}`
      const quickRange = getQuickRange(label, config)

      if (!quickRange?.from) {
        continue
      }

      const fromDeltaMs = Math.abs(range.from.getTime() - quickRange.from.getTime())

      if (fromDeltaMs <= 60_000) {
        return label
      }
    }
  }

  return undefined
}

export const TimeRangeSelector: React.FC<TimeRangeSelectorProps> = ({
  onChange,
  onClear,
  defaultRange,
  defaultSelectedQuickRange = 'Last 1 hour',
  className,
}) => {
  const pickerRef = React.useRef<DateRangePickerRef>(null)
  const fallbackRange = useMemo(
    () => getQuickRange(defaultSelectedQuickRange, quickRanges) ?? { from: subHours(new Date(), 1), to: new Date() },
    [defaultSelectedQuickRange],
  )
  const defaultFromTime = defaultRange?.from.getTime()
  const defaultToTime = defaultRange?.to.getTime()
  const selectedQuickRange = useMemo(
    () =>
      defaultFromTime != null && defaultToTime != null
        ? getQuickRangeLabel({ from: new Date(defaultFromTime), to: new Date(defaultToTime) }, quickRanges)
        : undefined,
    [defaultFromTime, defaultToTime],
  )
  const [dateRange, setDateRange] = useState<DateRange>(() => {
    if (defaultRange) {
      return { from: defaultRange.from, to: defaultRange.to }
    }
    return fallbackRange
  })
  const [isAllTimeSelected, setIsAllTimeSelected] = useState(false)

  useEffect(() => {
    if (defaultFromTime != null && defaultToTime != null) {
      setIsAllTimeSelected(false)
      setDateRange({ from: new Date(defaultFromTime), to: new Date(defaultToTime) })
      return
    }

    if (isAllTimeSelected) {
      setDateRange({ from: undefined, to: undefined })
      return
    }

    setDateRange(fallbackRange)
  }, [defaultFromTime, defaultToTime, fallbackRange, isAllTimeSelected])

  const handleChange = (range: DateRange) => {
    setDateRange(range)
    if (range.from && range.to) {
      setIsAllTimeSelected(false)
      onChange(range.from, range.to)
      return
    }

    setIsAllTimeSelected(true)
    onClear?.()
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
      selectedQuickRange={selectedQuickRange}
      defaultSelectedQuickRange={defaultSelectedQuickRange}
    />
  )
}
