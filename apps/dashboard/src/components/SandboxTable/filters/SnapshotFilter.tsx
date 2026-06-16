/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Command,
  CommandCheckboxItem,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandInputButton,
  CommandList,
} from '@/components/ui/command'
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
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { SnapshotDto } from '@daytona/api-client'
import { Camera } from 'lucide-react'

interface SnapshotFilterProps {
  value: string[]
  onFilterChange: (value: string[] | undefined) => void
  snapshots: SnapshotDto[]
  snapshotsDataIsLoading: boolean
  snapshotsDataHasMore?: boolean
  onChangeSnapshotSearchValue?: (name?: string) => void
}

export function SnapshotFilterIndicator({
  value,
  onFilterChange,
  snapshots,
  snapshotsDataIsLoading,
  snapshotsDataHasMore,
  onChangeSnapshotSearchValue,
}: SnapshotFilterProps) {
  const selectedSnapshots = value.map((snapshotName) => ({
    value: snapshotName,
    label: snapshotName,
  }))

  return (
    <FacetedFilterRoot title="Snapshot" hasValue={value.length > 0} onClear={() => onFilterChange(undefined)}>
      <FacetedFilterAnchor>
        <FacetedFilterLabelTrigger icon={<Camera />} aria-label="Filter by Snapshot">
          Snapshot
        </FacetedFilterLabelTrigger>
        <FacetedFilterOperator />
        <FacetedFilterValueTrigger
          className={cn({
            'px-1': value.length <= 1,
            'px-2': value.length > 1,
          })}
          aria-label="Edit Snapshot filter"
        >
          <FacetedFilterValues title="Snapshot" items={selectedSnapshots} maxValues={1} />
        </FacetedFilterValueTrigger>
        <FacetedFilterClear aria-label="Clear Snapshot filter" />
      </FacetedFilterAnchor>
      <FacetedFilterContent className="p-0 w-[240px]">
        <SnapshotFilter
          value={value}
          onFilterChange={onFilterChange}
          snapshots={snapshots}
          snapshotsDataIsLoading={snapshotsDataIsLoading}
          snapshotsDataHasMore={snapshotsDataHasMore}
          onChangeSnapshotSearchValue={onChangeSnapshotSearchValue}
        />
      </FacetedFilterContent>
    </FacetedFilterRoot>
  )
}

export function SnapshotFilter({
  value,
  onFilterChange,
  snapshots,
  snapshotsDataIsLoading,
  snapshotsDataHasMore,
  onChangeSnapshotSearchValue,
}: SnapshotFilterProps) {
  const handleSelect = (snapshotName: string) => {
    const newValue = value.includes(snapshotName)
      ? value.filter((name) => name !== snapshotName)
      : [...value, snapshotName]
    onFilterChange(newValue.length > 0 ? newValue : undefined)
  }

  return (
    <Command>
      <CommandInput placeholder="Search...">
        <CommandInputButton
          className="text-sm text-muted-foreground hover:text-primary px-2"
          onClick={() => onFilterChange(undefined)}
        >
          Clear
        </CommandInputButton>
      </CommandInput>

      <CommandList>
        {snapshotsDataIsLoading ? (
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
                <CommandCheckboxItem
                  key={snapshot.id}
                  onSelect={() => handleSelect(snapshot.name ?? '')}
                  value={snapshot.name}
                  className="cursor-pointer"
                  checked={value.includes(snapshot.name ?? '')}
                >
                  {snapshot.name}
                </CommandCheckboxItem>
              ))}
            </CommandGroup>
            {snapshotsDataHasMore && (
              <div className="p-2 text-xs text-muted-foreground text-center">Search to load more results</div>
            )}
          </>
        )}
      </CommandList>
    </Command>
  )
}
