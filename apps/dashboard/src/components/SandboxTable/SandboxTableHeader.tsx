/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Calendar, Camera, Columns, Cpu, Globe, HardDrive, ListFilter, MemoryStick, Square, Tag } from 'lucide-react'
import { SearchInput } from '../SearchInput'
import { TableColumnVisibilityToggle } from '../TableColumnVisibilityToggle'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
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
  labelOptions,
  regionOptions,
  regionsDataIsLoading,
  snapshots,
  loadingSnapshots,
}: SandboxTableHeaderProps) {
  const hasStateFilter = ((table.getColumn('state')?.getFilterValue() as string[]) || []).length > 0
  const hasSnapshotFilter = ((table.getColumn('snapshot')?.getFilterValue() as string[]) || []).length > 0
  const hasRegionFilter = ((table.getColumn('region')?.getFilterValue() as string[]) || []).length > 0
  const hasLabelsFilter = ((table.getColumn('labels')?.getFilterValue() as string[]) || []).length > 0
  const hasLastEventFilter = ((table.getColumn('lastEvent')?.getFilterValue() as Date[]) || []).length > 0
  const hasResourceFilter = RESOURCE_FILTERS.some(({ type }) => {
    return Boolean((table.getColumn('resources')?.getFilterValue() as ResourceFilterValue | undefined)?.[type])
  })

  const hasActiveFilters =
    hasStateFilter || hasSnapshotFilter || hasRegionFilter || hasLabelsFilter || hasLastEventFilter || hasResourceFilter

  const handleChangeFilter = (value: string) => {
    table.setGlobalFilter(value)
    table.setPageIndex(0)
  }

  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <div className="flex flex-1 items-center gap-2 min-w-0">
          <SearchInput
            debounced
            value={(table.getState().globalFilter as string) ?? ''}
            onValueChange={handleChangeFilter}
            placeholder="Search by Name, ID, Snapshot, Region, or Label"
            containerClassName="min-w-0 flex-1 sm:max-w-sm"
          />

          <DropdownMenu modal={false}>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="shrink-0">
                <ListFilter className="w-4 h-4" />
                Filter
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-40" align="start">
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
                      onFilterChange={(value: Date[] | undefined) =>
                        table.getColumn('lastEvent')?.setFilterValue(value)
                      }
                      value={(table.getColumn('lastEvent')?.getFilterValue() as Date[]) || []}
                    />
                  </DropdownMenuSubContent>
                </DropdownMenuPortal>
              </DropdownMenuSub>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <div className="hidden sm:block ml-auto">
          <DropdownMenu modal={false}>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="shrink-0">
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
        </div>
      </div>

      {hasActiveFilters ? (
        <div className="flex h-8 items-center gap-1 overflow-x-auto scrollbar-hide">
          {hasStateFilter && (
            <StateFilterIndicator
              value={(table.getColumn('state')?.getFilterValue() as string[]) || []}
              onFilterChange={(value) => table.getColumn('state')?.setFilterValue(value)}
            />
          )}

          {hasSnapshotFilter && (
            <SnapshotFilterIndicator
              value={(table.getColumn('snapshot')?.getFilterValue() as string[]) || []}
              onFilterChange={(value) => table.getColumn('snapshot')?.setFilterValue(value)}
              snapshots={snapshots}
              loadingSnapshots={loadingSnapshots}
            />
          )}

          {hasRegionFilter && (
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

          {hasLabelsFilter && (
            <LabelFilterIndicator
              value={(table.getColumn('labels')?.getFilterValue() as string[]) || []}
              onFilterChange={(value) => table.getColumn('labels')?.setFilterValue(value)}
              options={labelOptions}
            />
          )}

          {hasLastEventFilter && (
            <LastEventFilterIndicator
              value={(table.getColumn('lastEvent')?.getFilterValue() as Date[]) || []}
              onFilterChange={(value) => table.getColumn('lastEvent')?.setFilterValue(value)}
            />
          )}
        </div>
      ) : null}
    </div>
  )
}
