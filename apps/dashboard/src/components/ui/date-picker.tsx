/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { format } from 'date-fns'
import { Calendar as CalendarIcon, X } from 'lucide-react'

import { Calendar } from '@/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { InputGroup, InputGroupAddon, InputGroupButton } from './input-group'

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
      <InputGroup className="overflow-clip">
        <PopoverTrigger
          data-slot="input-group-control"
          id={id}
          data-empty={!value}
          className="flex flex-1 data-[empty=true]:text-muted-foreground justify-between text-left font-normal focus-visible:outline-none pl-2 h-full"
        >
          <div className="flex items-center gap-2">
            <CalendarIcon className="h-4 w-4" />
            {value ? format(value, 'PPP') : <span>Select date</span>}
          </div>
        </PopoverTrigger>
        {value && (
          <InputGroupAddon align="inline-end">
            <InputGroupButton type="button" size="icon-xs" onClick={handleClear}>
              <X className="h-3 w-3" />
            </InputGroupButton>
          </InputGroupAddon>
        )}
      </InputGroup>

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
