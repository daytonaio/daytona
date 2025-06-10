/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { format } from 'date-fns'
import { Calendar as CalendarIcon, X } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'

interface DatePickerProps {
  value?: Date
  onChange: (date?: Date) => void
  required?: boolean
  disabledBefore?: Date
  id?: string
}

export function DatePicker({ value, onChange, required, disabledBefore, id }: DatePickerProps) {
  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation()
    onChange(undefined)
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          id={id}
          variant="outline"
          data-empty={!value}
          className="flex w-full data-[empty=true]:text-muted-foreground justify-between text-left font-normal hover:bg-background"
        >
          <div className="flex items-center gap-2">
            <CalendarIcon className="h-4 w-4" />
            {value ? format(value, 'PPP') : <span>Select date</span>}
          </div>
          {value && (
            <Button type="button" variant="ghost" size="sm" className="h-auto p-1 hover:bg-muted" onClick={handleClear}>
              <X className="h-3 w-3" />
            </Button>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          selected={value}
          onSelect={onChange}
          required={required}
          disabled={(() => {
            const conditions = []
            if (disabledBefore) {
              conditions.push({ before: disabledBefore })
            }
            return conditions
          })()}
        />
      </PopoverContent>
    </Popover>
  )
}
