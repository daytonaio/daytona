/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandInputButton,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { SnapshotDto } from '@daytonaio/api-client'
import { Check, Loader2, X } from 'lucide-react'
import { useState } from 'react'

interface SnapshotFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
  snapshots: SnapshotDto[]
  isLoading: boolean
  hasMore?: boolean
  onChangeSnapshotSearchValue: (name?: string) => void
}

export function SnapshotFilterIndicator({
  value,
  onFilterChange,
  snapshots,
  isLoading,
  hasMore,
  onChangeSnapshotSearchValue,
}: SnapshotFilterProps) {
  return (
    <div className="flex items-center h-6 gap-0.5 rounded-sm border border-border bg-muted/80 hover:bg-muted/50 text-sm">
      <Popover>
        <PopoverTrigger className="max-w-[160px] overflow-hidden text-ellipsis whitespace-nowrap text-muted-foreground px-2">
          Snapshot: <span className="text-primary font-medium">{value.length} selected</span>
        </PopoverTrigger>

        <PopoverContent className="p-0 w-[240px]" align="start">
          <SnapshotFilter
            value={value}
            onFilterChange={onFilterChange}
            snapshots={snapshots}
            isLoading={isLoading}
            hasMore={hasMore}
            onChangeSnapshotSearchValue={onChangeSnapshotSearchValue}
          />
        </PopoverContent>
      </Popover>

      <button className="h-6 w-5 p-0 border-0 hover:text-muted-foreground" onClick={() => onFilterChange(undefined)}>
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}

export function SnapshotFilter({
  value,
  onFilterChange,
  snapshots,
  isLoading,
  hasMore,
  onChangeSnapshotSearchValue,
}: SnapshotFilterProps) {
  const [searchValue, setSearchValue] = useState('')

  const handleSelect = (snapshotName: string) => {
    const newValue = value.includes(snapshotName)
      ? value.filter((name) => name !== snapshotName)
      : [...value, snapshotName]
    onFilterChange(newValue.length > 0 ? newValue : undefined)
  }

  const handleSearchChange = (search: string | number) => {
    const searchStr = String(search)
    setSearchValue(searchStr)
    if (onChangeSnapshotSearchValue) {
      onChangeSnapshotSearchValue(searchStr || undefined)
    }
  }

  return (
    <Command>
      <CommandInput placeholder="Search..." className="" value={searchValue} onValueChange={setSearchValue}>
        <CommandInputButton
          onClick={() => {
            onFilterChange(undefined)
            setSearchValue('')
            if (onChangeSnapshotSearchValue) {
              onChangeSnapshotSearchValue(undefined)
            }
          }}
        >
          Clear
        </CommandInputButton>
      </CommandInput>
      {hasMore && (
        <div className="px-2 pb-2 mt-2">
          <div className="text-xs text-muted-foreground bg-muted/50 rounded px-2 py-1">
            Please refine your search to see more Snapshots.
          </div>
        </div>
      )}
      <CommandList>
        {isLoading ? (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="h-4 w-4 animate-spin mr-2" />
            <span className="text-sm text-muted-foreground">Loading Snapshots...</span>
          </div>
        ) : (
          <>
            <CommandEmpty>No Snapshots found.</CommandEmpty>
            <CommandGroup>
              {snapshots.map((snapshot) => (
                <CommandItem
                  key={snapshot.id}
                  onSelect={() => handleSelect(snapshot.name ?? '')}
                  value={snapshot.name}
                  className="cursor-pointer"
                >
                  <div
                    className={cn(
                      'mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary',
                      value.includes(snapshot.name ?? '')
                        ? 'bg-primary text-primary-foreground'
                        : 'opacity-50 [&_svg]:invisible',
                    )}
                  >
                    <Check className={cn('h-4 w-4')} />
                  </div>
                  <span>{snapshot.name}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}
      </CommandList>
    </Command>
  )
}
