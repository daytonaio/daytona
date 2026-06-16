/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import {
  FacetedFilterAnchor,
  FacetedFilterClear,
  FacetedFilterContent,
  FacetedFilterLabelTrigger,
  FacetedFilterOperator,
  FacetedFilterRoot,
  FacetedFilterValueTrigger,
  FacetedFilterValues,
} from '@/components/ui/faceted-filter'
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { Calendar } from '@/components/ui/calendar'
import { Label } from '@/components/ui/label'
import { useState } from 'react'
import { format } from 'date-fns'
import { CalendarIcon, Calendar as LastEventIcon } from 'lucide-react'

interface LastEventFilterProps {
  value: (Date | undefined)[]
  onFilterChange: (value: (Date | undefined)[] | undefined) => void
}

export function LastEventFilterIndicator({ value, onFilterChange }: LastEventFilterProps) {
  const selectedDateLabel = value
    .filter((date): date is Date => date !== undefined)
    .map((date) => format(date, 'PPP'))
    .join(' - ')
  const selectedDates = selectedDateLabel
    ? [
        {
          value: selectedDateLabel,
          label: selectedDateLabel,
        },
      ]
    : []

  return (
    <FacetedFilterRoot title="Last Event" hasValue={selectedDates.length > 0} onClear={() => onFilterChange(undefined)}>
      <FacetedFilterAnchor>
        <FacetedFilterLabelTrigger icon={<LastEventIcon />} aria-label="Filter by Last Event">
          Last Event
        </FacetedFilterLabelTrigger>
        <FacetedFilterOperator />
        <FacetedFilterValueTrigger className="px-1" aria-label="Edit Last Event filter">
          <FacetedFilterValues title="Last Event" items={selectedDates} maxValues={1} />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label="Clear Last Event filter" />
      </FacetedFilterAnchor>
      <FacetedFilterContent className="p-3 w-auto">
        <LastEventFilter onFilterChange={onFilterChange} value={value} />
      </FacetedFilterContent>
    </FacetedFilterRoot>
  )
}

interface LastEventFilterContentProps {
  onFilterChange: (value: (Date | undefined)[] | undefined) => void
  value: (Date | undefined)[]
}

export function LastEventFilter({ onFilterChange, value }: LastEventFilterContentProps) {
  const [fromDate, setFromDate] = useState<Date | undefined>(value[0])
  const [toDate, setToDate] = useState<Date | undefined>(value[1])

  const handleFromDateSelect = (selectedDate: Date | undefined) => {
    setFromDate(selectedDate)
    const dates = [selectedDate, toDate]
    const hasAnyDate = dates.some((date) => date !== undefined)
    onFilterChange(hasAnyDate ? dates : undefined)
  }

  const handleToDateSelect = (selectedDate: Date | undefined) => {
    setToDate(selectedDate)
    const dates = [fromDate, selectedDate]
    const hasAnyDate = dates.some((date) => date !== undefined)
    onFilterChange(hasAnyDate ? dates : undefined)
  }

  const handleClear = () => {
    setFromDate(undefined)
    setToDate(undefined)
    onFilterChange(undefined)
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <Label>Last event</Label>
        <button className="text-sm text-muted-foreground hover:text-primary px-2" onClick={() => handleClear()}>
          Clear
        </button>
      </div>
      <div className="flex gap-2 items-center">
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline" className={cn('min-w-40', { 'text-muted-foreground': !fromDate })}>
              <CalendarIcon className=" h-4 w-4" />
              {fromDate ? format(fromDate, 'PPP') : <span>Pick a date</span>}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0" align="start">
            <Calendar mode="single" selected={fromDate} onSelect={handleFromDateSelect} initialFocus />
          </PopoverContent>
        </Popover>

        <div className="w-4 flex-shrink-0 h-[1px] bg-border"></div>

        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline" className={cn('min-w-40', { 'text-muted-foreground': !toDate })}>
              <CalendarIcon className=" h-4 w-4" />
              {toDate ? format(toDate, 'PPP') : <span>Pick a date</span>}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0" align="start">
            <Calendar mode="single" selected={toDate} onSelect={handleToDateSelect} initialFocus />
          </PopoverContent>
        </Popover>
      </div>
    </div>
  )
}
