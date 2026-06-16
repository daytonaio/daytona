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
import { Input } from '@/components/ui/input'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { Plus, Tag, Trash2 } from 'lucide-react'
import { useState } from 'react'

interface LabelFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
}

export function LabelFilterIndicator({ value, onFilterChange }: Pick<LabelFilterProps, 'value' | 'onFilterChange'>) {
  const selectedLabels = value.map((label) => ({
    value: label,
    label,
  }))

  return (
    <FacetedFilterRoot title="Label" hasValue={value.length > 0} onClear={() => onFilterChange(undefined)}>
      <FacetedFilterAnchor>
        <FacetedFilterLabelTrigger icon={<Tag />} aria-label="Filter by Label">
          Labels
        </FacetedFilterLabelTrigger>
        <FacetedFilterOperator />
        <FacetedFilterValueTrigger
          className={cn({
            'px-1': value.length <= 1,
            'px-2': value.length > 1,
          })}
          aria-label="Edit Label filter"
        >
          <FacetedFilterValues title="Label" items={selectedLabels} maxValues={1} />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label="Clear Label filter" />
      </FacetedFilterAnchor>
      <FacetedFilterContent className="p-0 w-[320px]">
        <LabelFilter value={value} onFilterChange={onFilterChange} />
      </FacetedFilterContent>
    </FacetedFilterRoot>
  )
}

export function LabelFilter({ value, onFilterChange }: LabelFilterProps) {
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')

  const labelPairs = value.map((labelString) => {
    const [key, ...valueParts] = labelString.split(': ')
    return { key: key || '', value: valueParts.join(': ') || '' }
  })

  const addKeyValuePair = () => {
    if (newKey.trim() && newValue.trim()) {
      const newLabelString = `${newKey.trim()}: ${newValue.trim()}`
      const updatedValue = [...value, newLabelString]
      onFilterChange(updatedValue)
      setNewKey('')
      setNewValue('')
    }
  }

  const removeKeyValuePair = (index: number) => {
    const updatedValue = value.filter((_, i) => i !== index)
    onFilterChange(updatedValue.length > 0 ? updatedValue : undefined)
  }

  const clearAll = () => {
    onFilterChange(undefined)
  }

  return (
    <div className="p-3 space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium">Labels</h4>
        <button className="text-sm text-muted-foreground hover:text-primary pl-2" onClick={clearAll}>
          Clear
        </button>
      </div>

      {labelPairs.length > 0 && (
        <div className="space-y-2">
          <div className="space-y-1 max-h-32 overflow-y-auto">
            {labelPairs.map((pair, index) => (
              <div key={index} className="flex items-center gap-2 p-2 bg-muted/50 rounded-sm">
                <div className="flex-1 flex items-center gap-1 text-sm min-w-0">
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div className="truncate flex-shrink-0 max-w-[50%] rounded-sm bg-blue-100 dark:bg-blue-950 text-blue-800 dark:text-blue-200 px-1 cursor-default">
                        {pair.key}
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="max-w-[300px] break-words">{pair.key}</p>
                    </TooltipContent>
                  </Tooltip>

                  <Tooltip>
                    <TooltipTrigger asChild>
                      <span className="truncate flex-1 text-muted-foreground cursor-default">{pair.value}</span>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="max-w-[300px] break-words">{pair.value}</p>
                    </TooltipContent>
                  </Tooltip>
                </div>
                <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => removeKeyValuePair(index)}>
                  <Trash2 className="h-3 w-3" />
                </Button>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="space-y-2">
        <div className="space-y-2">
          <Input
            placeholder="Key"
            value={newKey}
            onChange={(e) => setNewKey(e.target.value)}
            className="h-8"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && newKey.trim() && newValue.trim()) {
                addKeyValuePair()
              }
            }}
          />
          <Input
            placeholder="Value"
            value={newValue}
            onChange={(e) => setNewValue(e.target.value)}
            className="h-8"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && newKey.trim() && newValue.trim()) {
                addKeyValuePair()
              }
            }}
          />
          <Button
            variant="outline"
            size="sm"
            className="w-full h-8"
            onClick={addKeyValuePair}
            disabled={!newKey.trim() || !newValue.trim()}
          >
            <Plus className="h-3 w-3 mr-1" />
            Add Label
          </Button>
        </div>
      </div>
    </div>
  )
}
