/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import {
  ArrowUpDown,
  Calendar,
  Camera,
  Check,
  Columns,
  Cpu,
  Globe,
  HardDrive,
  ListFilter,
  MemoryStick,
  RefreshCw,
  Square,
  Tag,
} from 'lucide-react'
import * as React from 'react'
import { DebouncedInput } from '../DebouncedInput'
import { TableColumnVisibilityToggle } from '../TableColumnVisibilityToggle'
import { Button } from '../ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandInputButton,
  CommandItem,
  CommandList,
} from '../ui/command'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import { Popover, PopoverContent, PopoverTrigger } from '../ui/popover'
import { LabelFilter, LabelFilterIndicator } from './filters/LabelFilter'
import { LastEventFilter, LastEventFilterIndicator } from './filters/LastEventFilter'
import { RegionFilter, RegionFilterIndicator } from './filters/RegionFilter'
import { ResourceFilter, ResourceFilterIndicator, ResourceFilterValue } from './filters/ResourceFilter'
import { SnapshotFilter, SnapshotFilterIndicator } from './filters/SnapshotFilter'
import { StateFilter, StateFilterIndicator } from './filters/StateFilter'
import { SandboxTableHeaderProps } from './types'

const RESOURCE_FILTERS = [
  { type: 'cpu' as const, label: 'CPU', icon: Cpu },
  { type: 'memory' as const, label: 'Memory', icon: MemoryStick },
  { type: 'disk' as const, label: 'Disk', icon: HardDrive },
]

export function SandboxTableHeader({
  table,
  regionOptions,
  regionsDataIsLoading,
  snapshots,
  snapshotsDataIsLoading,
  snapshotsDataHasMore,
  onChangeSnapshotSearchValue,
  onRefresh,
  isRefreshing = false,
}: SandboxTableHeaderProps) {
  const [open, setOpen] = React.useState(false)
  const currentSort = table.getState().sorting[0]?.id || ''

  const sortableColumns = [
    { id: 'name', label: 'Name' },
    { id: 'state', label: 'State' },
    { id: 'snapshot', label: 'Snapshot' },
    { id: 'region', label: 'Region' },
    { id: 'lastEvent', label: 'Last Event' },
  ]

  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
      <div className="flex flex-wrap gap-2 items-center">
        <DebouncedInput
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(value) => table.getColumn('name')?.setFilterValue(value)}
          placeholder="Search by Name or UUID"
          className="w-[240px]"
        />

        <Button variant="outline" onClick={onRefresh} disabled={isRefreshing} className="flex items-center gap-2">
          <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          Refresh
        </Button>

        <DropdownMenu modal={false}>
          <DropdownMenuTrigger asChild>
            <Button variant="outline">
              <Columns className="w-4 h-4" />
              View
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-[200px] p-0">
            <TableColumnVisibilityToggle
              columns={table.getAllColumns().filter((column) => ['name', 'id', 'labels'].includes(column.id))}
              getColumnLabel={(id: string) => {
                switch (id) {
                  case 'name':
                    return 'Name'
                  case 'id':
                    return 'UUID'
                  case 'labels':
                    return 'Labels'
                  default:
                    return id
                }
              }}
            />
          </DropdownMenuContent>
        </DropdownMenu>

        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button variant="outline" role="combobox" aria-expanded={open} className="justify-between">
              {currentSort ? (
                <div className="flex items-center gap-2">
                  <div className="text-muted-foreground font-normal">
                    Sorted by:{' '}
                    <span className="font-medium text-primary">
                      {sortableColumns.find((column) => column.id === currentSort)?.label}
                    </span>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <ArrowUpDown className="w-4 h-4" />
                  <span>Sort</span>
                </div>
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[240px] p-0" align="start">
            <Command>
              <CommandInput placeholder="Search...">
                <CommandInputButton
                  aria-expanded={open}
                  className="justify-between"
                  onClick={() => {
                    table.resetSorting()
                    setOpen(false)
                  }}
                >
                  Reset
                </CommandInputButton>
              </CommandInput>
              <CommandList>
                <CommandEmpty>No column found.</CommandEmpty>
                <CommandGroup>
                  {sortableColumns.map((column) => (
                    <CommandItem
                      key={column.id}
                      value={column.id}
                      onSelect={(currentValue) => {
                        const col = table.getColumn(currentValue)
                        if (col) {
                          col.toggleSorting(false)
                        }
                        setOpen(false)
                      }}
                    >
                      <Check className={cn('mr-2 h-4 w-4', currentSort === column.id ? 'opacity-100' : 'opacity-0')} />
                      {column.label}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        <DropdownMenu modal={false}>
          <DropdownMenuTrigger asChild>
            <Button variant="outline">
              <ListFilter className="w-4 h-4" />
              Filter
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className=" w-40" align="start">
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>
                <Square className="w-4 h-4" />
                State
              </DropdownMenuSubTrigger>
              <DropdownMenuPortal>
                <DropdownMenuSubContent className="p-0 w-64">
                  <StateFilter
                    value={(table.getColumn('state')?.getFilterValue() as string[]) || []}
                    onFilterChange={(value) => table.getColumn('state')?.setFilterValue(value)}
                  />
                </DropdownMenuSubContent>
              </DropdownMenuPortal>
            </DropdownMenuSub>
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>
                <Camera className="w-4 h-4" />
                Snapshot
              </DropdownMenuSubTrigger>
              <DropdownMenuPortal>
                <DropdownMenuSubContent className="p-0 w-64">
                  <SnapshotFilter
                    value={(table.getColumn('snapshot')?.getFilterValue() as string[]) || []}
                    onFilterChange={(value) => table.getColumn('snapshot')?.setFilterValue(value)}
                    snapshots={snapshots}
                    isLoading={snapshotsDataIsLoading}
                    hasMore={snapshotsDataHasMore}
                    onChangeSnapshotSearchValue={onChangeSnapshotSearchValue}
                  />
                </DropdownMenuSubContent>
              </DropdownMenuPortal>
            </DropdownMenuSub>
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>
                <Globe className="w-4 h-4" />
                Region
              </DropdownMenuSubTrigger>
              <DropdownMenuPortal>
                <DropdownMenuSubContent className="p-0 w-64">
                  <RegionFilter
                    value={(table.getColumn('region')?.getFilterValue() as string[]) || []}
                    onFilterChange={(value) => table.getColumn('region')?.setFilterValue(value)}
                    options={regionOptions}
                    isLoading={regionsDataIsLoading}
                  />
                </DropdownMenuSubContent>
              </DropdownMenuPortal>
            </DropdownMenuSub>
            {RESOURCE_FILTERS.map(({ type, label, icon: Icon }) => (
              <DropdownMenuSub key={type}>
                <DropdownMenuSubTrigger>
                  <Icon className="w-4 h-4" />
                  {label}
                </DropdownMenuSubTrigger>
                <DropdownMenuPortal>
                  <DropdownMenuSubContent className="p-3 w-64">
                    <ResourceFilter
                      value={(table.getColumn('resources')?.getFilterValue() as ResourceFilterValue) || {}}
                      onFilterChange={(value) => table.getColumn('resources')?.setFilterValue(value)}
                      resourceType={type}
                    />
                  </DropdownMenuSubContent>
                </DropdownMenuPortal>
              </DropdownMenuSub>
            ))}
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>
                <Tag className="w-4 h-4" />
                Labels
              </DropdownMenuSubTrigger>
              <DropdownMenuPortal>
                <DropdownMenuSubContent className="p-0 w-64">
                  <LabelFilter
                    value={(table.getColumn('labels')?.getFilterValue() as string[]) || []}
                    onFilterChange={(value) => table.getColumn('labels')?.setFilterValue(value)}
                  />
                </DropdownMenuSubContent>
              </DropdownMenuPortal>
            </DropdownMenuSub>
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>
                <Calendar className="w-4 h-4" />
                Last Event
              </DropdownMenuSubTrigger>
              <DropdownMenuPortal>
                <DropdownMenuSubContent className="p-3 w-92">
                  <LastEventFilter
                    onFilterChange={(value) => table.getColumn('lastEvent')?.setFilterValue(value)}
                    value={(table.getColumn('lastEvent')?.getFilterValue() as Date[]) || []}
                  />
                </DropdownMenuSubContent>
              </DropdownMenuPortal>
            </DropdownMenuSub>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className="flex flex-1 gap-1 overflow-x-auto scrollbar-hide h-8 items-center">
        {(table.getColumn('state')?.getFilterValue() as string[])?.length > 0 && (
          <StateFilterIndicator
            value={(table.getColumn('state')?.getFilterValue() as string[]) || []}
            onFilterChange={(value) => table.getColumn('state')?.setFilterValue(value)}
          />
        )}

        {(table.getColumn('snapshot')?.getFilterValue() as string[])?.length > 0 && (
          <SnapshotFilterIndicator
            value={(table.getColumn('snapshot')?.getFilterValue() as string[]) || []}
            onFilterChange={(value) => table.getColumn('snapshot')?.setFilterValue(value)}
            snapshots={snapshots}
            isLoading={snapshotsDataIsLoading}
            hasMore={snapshotsDataHasMore}
            onChangeSnapshotSearchValue={onChangeSnapshotSearchValue}
          />
        )}

        {(table.getColumn('region')?.getFilterValue() as string[])?.length > 0 && (
          <RegionFilterIndicator
            value={(table.getColumn('region')?.getFilterValue() as string[]) || []}
            onFilterChange={(value) => table.getColumn('region')?.setFilterValue(value)}
            options={regionOptions}
            isLoading={regionsDataIsLoading}
          />
        )}

        {RESOURCE_FILTERS.map(({ type }) => {
          const resourceValue = (table.getColumn('resources')?.getFilterValue() as ResourceFilterValue)?.[type]
          return resourceValue ? (
            <ResourceFilterIndicator
              key={type}
              value={table.getColumn('resources')?.getFilterValue() as ResourceFilterValue}
              onFilterChange={(value) => table.getColumn('resources')?.setFilterValue(value)}
              resourceType={type}
            />
          ) : null
        })}

        {(table.getColumn('labels')?.getFilterValue() as string[])?.length > 0 && (
          <LabelFilterIndicator
            value={(table.getColumn('labels')?.getFilterValue() as string[]) || []}
            onFilterChange={(value) => table.getColumn('labels')?.setFilterValue(value)}
          />
        )}

        {(table.getColumn('lastEvent')?.getFilterValue() as Date[])?.length > 0 && (
          <LastEventFilterIndicator
            value={(table.getColumn('lastEvent')?.getFilterValue() as Date[]) || []}
            onFilterChange={(value) => table.getColumn('lastEvent')?.setFilterValue(value)}
          />
        )}
      </div>
    </div>
  )
}
