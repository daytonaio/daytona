/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import * as React from 'react'
import {
  ArrowUpDown,
  Check,
  ListFilter,
  Square,
  Globe,
  Cpu,
  Tag,
  Calendar,
  Camera,
  HardDrive,
  MemoryStick,
} from 'lucide-react'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuPortal,
} from '../ui/dropdown-menu'
import { DebouncedInput } from '../DebouncedInput'
import { SandboxTableHeaderProps } from './types'
import { StateFilter, StateFilterIndicator } from './filters/StateFilter'
import { RegionFilter, RegionFilterIndicator } from './filters/RegionFilter'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '../ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '../ui/popover'
import { cn } from '@/lib/utils'
import { LastEventFilter, LastEventFilterIndicator } from './filters/LastEventFilter'
import { SnapshotFilter, SnapshotFilterIndicator } from './filters/SnapshotFilter'
import { ResourceFilter, ResourceFilterIndicator, ResourceFilterValue } from './filters/ResourceFilter'
import { LabelFilter, LabelFilterIndicator } from './filters/LabelFilter'

const RESOURCE_FILTERS = [
  { type: 'cpu' as const, label: 'CPU', icon: Cpu },
  { type: 'memory' as const, label: 'Memory', icon: MemoryStick },
  { type: 'disk' as const, label: 'Disk', icon: HardDrive },
]

export function SandboxTableHeader({
  table,
  labelOptions,
  regionOptions,
  snapshots,
  loadingSnapshots,
}: SandboxTableHeaderProps) {
  const [open, setOpen] = React.useState(false)
  const currentSort = table.getState().sorting[0]?.id || ''

  const sortableColumns = [
    { id: 'id', label: 'ID' },
    { id: 'state', label: 'State' },
    { id: 'snapshot', label: 'Snapshot' },
    { id: 'region', label: 'Region' },
    { id: 'lastEvent', label: 'Last Event' },
  ]

  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center mb-4">
      <div className="flex gap-2 items-center">
        <DebouncedInput
          value={(table.getColumn('id')?.getFilterValue() as string) ?? ''}
          onChange={(value) => table.getColumn('id')?.setFilterValue(value)}
          placeholder="Search by ID"
          className="max-w-[200px]"
        />

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
              <div className="flex items-center gap-2  px-2 pt-2 pb-1">
                <CommandInput placeholder="Search..." className="border border-border rounded-md h-8" />

                <Button
                  variant="link"
                  aria-expanded={open}
                  className="justify-between"
                  onClick={() => {
                    table.resetSorting()
                    setOpen(false)
                  }}
                >
                  Reset
                </Button>
              </div>
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
                    loadingSnapshots={loadingSnapshots}
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
                    options={labelOptions}
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
                    onFilterChange={(value: Date[] | undefined) => table.getColumn('lastEvent')?.setFilterValue(value)}
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
            loadingSnapshots={loadingSnapshots}
          />
        )}

        {(table.getColumn('region')?.getFilterValue() as string[])?.length > 0 && (
          <RegionFilterIndicator
            value={(table.getColumn('region')?.getFilterValue() as string[]) || []}
            onFilterChange={(value) => table.getColumn('region')?.setFilterValue(value)}
            options={regionOptions}
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
            options={labelOptions}
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
