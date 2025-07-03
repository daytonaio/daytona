/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SnapshotDto } from '@daytonaio/api-client'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { Skeleton } from '@/components/ui/skeleton'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { Check } from 'lucide-react'
import { X } from 'lucide-react'

interface SnapshotFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
  snapshots: SnapshotDto[]
  loadingSnapshots: boolean
}

export function SnapshotFilterIndicator({ value, onFilterChange, snapshots, loadingSnapshots }: SnapshotFilterProps) {
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
            loadingSnapshots={loadingSnapshots}
          />
        </PopoverContent>
      </Popover>

      <button className="h-6 w-5 p-0 border-0 hover:text-muted-foreground" onClick={() => onFilterChange(undefined)}>
        <X className="h-3 w-3" />
      </button>
    </div>
  )
}

export function SnapshotFilter({ value, onFilterChange, snapshots, loadingSnapshots }: SnapshotFilterProps) {
  const handleSelect = (snapshotName: string) => {
    const newValue = value.includes(snapshotName)
      ? value.filter((name) => name !== snapshotName)
      : [...value, snapshotName]
    onFilterChange(newValue.length > 0 ? newValue : undefined)
  }

  return (
    <Command>
      <div className="flex items-center gap-2 justify-between p-2">
        <CommandInput placeholder="Filter by snapshot..." className="border border-border rounded-md h-8" />
        <button
          className="text-sm text-muted-foreground hover:text-primary px-2"
          onClick={() => onFilterChange(undefined)}
        >
          Clear
        </button>
      </div>
      <CommandList>
        {loadingSnapshots ? (
          <div className="p-1">
            <Skeleton className="h-8 w-full mb-1" />
            <Skeleton className="h-8 w-full mb-1" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : (
          <>
            <CommandEmpty>No snapshots found.</CommandEmpty>
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
