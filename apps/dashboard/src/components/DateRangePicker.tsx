'use client'

import { useState, useEffect, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { ChevronLeft, ChevronRight } from 'lucide-react'

interface DateRangePickerProps {
  onRangeSelect: (start: string, end: string) => void
  onClose: () => void
}

export function DateRangePicker({ onRangeSelect, onClose }: DateRangePickerProps) {
  const [currentMonth, setCurrentMonth] = useState(new Date())
  const [selectedStart, setSelectedStart] = useState<Date | null>(null)
  const [selectedEnd, setSelectedEnd] = useState<Date | null>(null)
  const [position, setPosition] = useState<'left' | 'right'>('left')
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (containerRef.current) {
      const rect = containerRef.current.getBoundingClientRect()
      const viewportWidth = window.innerWidth
      const modalWidth = 384 // w-96 = 384px

      // If there's not enough space on the right, position to the left
      if (rect.right + modalWidth > viewportWidth) {
        setPosition('right')
      } else {
        setPosition('left')
      }
    }
  }, [])

  const presets = [
    { label: 'Week to date', days: 7 },
    { label: 'Month to date', days: 30 },
    { label: 'Last 7 days', days: 7 },
    { label: 'Last 14 days', days: 14 },
    { label: 'Last 30 days', days: 30 },
  ]

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('en-US', {
      month: '2-digit',
      day: '2-digit',
      year: '2-digit',
    })
  }

  const handlePresetClick = (days: number) => {
    const end = new Date()
    const start = new Date()
    start.setDate(start.getDate() - days)

    setSelectedStart(start)
    setSelectedEnd(end)
    onRangeSelect(formatDate(start), formatDate(end))
  }

  const getDaysInMonth = (date: Date) => {
    const year = date.getFullYear()
    const month = date.getMonth()
    const firstDay = new Date(year, month, 1)
    const lastDay = new Date(year, month + 1, 0)
    const daysInMonth = lastDay.getDate()
    const startingDayOfWeek = firstDay.getDay()

    const days = []

    // Add empty cells for days before the first day of the month
    for (let i = 0; i < startingDayOfWeek; i++) {
      days.push(null)
    }

    // Add days of the month
    for (let day = 1; day <= daysInMonth; day++) {
      days.push(new Date(year, month, day))
    }

    return days
  }

  const navigateMonth = (direction: 'prev' | 'next') => {
    setCurrentMonth((prev) => {
      const newMonth = new Date(prev)
      if (direction === 'prev') {
        newMonth.setMonth(newMonth.getMonth() - 1)
      } else {
        newMonth.setMonth(newMonth.getMonth() + 1)
      }
      return newMonth
    })
  }

  const handleDayClick = (date: Date) => {
    if (!selectedStart || (selectedStart && selectedEnd)) {
      setSelectedStart(date)
      setSelectedEnd(null)
    } else if (selectedStart && !selectedEnd) {
      if (date < selectedStart) {
        setSelectedStart(date)
        setSelectedEnd(selectedStart)
      } else {
        setSelectedEnd(date)
      }
    }
  }

  const isDateInRange = (date: Date) => {
    if (!selectedStart || !selectedEnd) return false
    return date >= selectedStart && date <= selectedEnd
  }

  const isDateSelected = (date: Date) => {
    return (
      (selectedStart && date.getTime() === selectedStart.getTime()) ||
      (selectedEnd && date.getTime() === selectedEnd.getTime())
    )
  }

  return (
    <Card
      ref={containerRef}
      className={`absolute top-full z-50 mt-2 p-0 w-96 bg-background border shadow-lg ${
        position === 'right' ? 'right-0' : 'left-0'
      }`}
    >
      <div className="flex">
        {/* Presets */}
        <div className="w-32 border-r border-border p-2 space-y-1">
          {presets.map((preset) => (
            <Button
              key={preset.label}
              variant="ghost"
              size="sm"
              className="w-full justify-start text-xs h-8 px-2"
              onClick={() => handlePresetClick(preset.days)}
            >
              {preset.label}
            </Button>
          ))}
        </div>

        {/* Calendar */}
        <div className="flex-1 p-4">
          <div className="flex items-center justify-between mb-4">
            <Button variant="ghost" size="sm" onClick={() => navigateMonth('prev')}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <div className="text-sm font-medium">
              {currentMonth.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
            </div>
            <Button variant="ghost" size="sm" onClick={() => navigateMonth('next')}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>

          {/* Days of week header */}
          <div className="grid grid-cols-7 gap-1 mb-2">
            {['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'].map((day) => (
              <div key={day} className="text-xs text-muted-foreground text-center p-1">
                {day}
              </div>
            ))}
          </div>

          {/* Calendar days */}
          <div className="grid grid-cols-7 gap-1">
            {getDaysInMonth(currentMonth).map((date, index) => (
              <div key={index} className="aspect-square">
                {date && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className={`w-full h-full text-xs p-0 ${
                      isDateSelected(date)
                        ? 'bg-primary text-primary-foreground'
                        : isDateInRange(date)
                          ? 'bg-primary/20'
                          : ''
                    }`}
                    onClick={() => handleDayClick(date)}
                  >
                    {date.getDate()}
                  </Button>
                )}
              </div>
            ))}
          </div>

          {/* Action buttons */}
          <div className="flex justify-between mt-4 pt-4 border-t">
            <Button variant="outline" size="sm" onClick={onClose}>
              Cancel
            </Button>
            <Button
              size="sm"
              disabled={!selectedStart || !selectedEnd}
              onClick={() => {
                if (selectedStart && selectedEnd) {
                  onRangeSelect(formatDate(selectedStart), formatDate(selectedEnd))
                  onClose()
                }
              }}
            >
              Apply
            </Button>
          </div>
        </div>
      </div>
    </Card>
  )
}
