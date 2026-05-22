/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Boxes,
  Calendar,
  Camera,
  Columns,
  Cpu,
  Globe,
  HardDrive,
  ListFilter,
  MemoryStick,
  Square,
  Tag,
} from 'lucide-react'
import { SearchInput } from '../SearchInput'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuPortal,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip'
import { LabelFilter, LabelFilterIndicator } from './filters/LabelFilter'
import { LastEventFilter, LastEventFilterIndicator } from './filters/LastEventFilter'
import { RegionFilter, RegionFilterIndicator } from './filters/RegionFilter'
import { ResourceFilter, ResourceFilterIndicator, ResourceFilterValue } from './filters/ResourceFilter'
import { SandboxClassFilter, SandboxClassFilterIndicator } from './filters/SandboxClassFilter'
import { SnapshotFilter, SnapshotFilterIndicator } from './filters/SnapshotFilter'
import { StateFilter, StateFilterIndicator } from './filters/StateFilter'
import { SandboxTableHeaderProps } from './types'

const RESOURCE_FILTERS = [
  { type: 'cpu' as const, label: 'CPU', icon: Cpu },
  { type: 'memory' as const, label: 'Memory', icon: MemoryStick },
  { type: 'disk' as const, label: 'Disk', icon: HardDrive },
]

const SANDBOX_TABLE_COLUMN_LABELS: Record<string, string> = {
  name: 'Name',
  id: 'UUID',
  state: 'State',
  class: 'Class',
  snapshot: 'Snapshot',
  region: 'Region',
  resources: 'Resources',
  labels: 'Labels',
  lastEvent: 'Last Event',
  createdAt: 'Created At',
}

export function SandboxTableHeader({
  table,
  labelOptions,
  regionOptions,
  regionsDataIsLoading,
  snapshots,
  loadingSnapshots,
}: SandboxTableHeaderProps) {
  const hasStateFilter = ((table.getColumn('state')?.getFilterValue() as string[]) || []).length > 0
  const hasClassFilter = ((table.getColumn('class')?.getFilterValue() as string[]) || []).length > 0
  const hasSnapshotFilter = ((table.getColumn('snapshot')?.getFilterValue() as string[]) || []).length > 0
  const hasRegionFilter = ((table.getColumn('region')?.getFilterValue() as string[]) || []).length > 0
  const hasLabelsFilter = ((table.getColumn('labels')?.getFilterValue() as string[]) || []).length > 0
  const hasLastEventFilter = ((table.getColumn('lastEvent')?.getFilterValue() as Date[]) || []).length > 0
  const hasResourceFilter = RESOURCE_FILTERS.some(({ type }) => {
    return Boolean((table.getColumn('resources')?.getFilterValue() as ResourceFilterValue | undefined)?.[type])
  })

  const hasActiveFilters =
    hasStateFilter ||
    hasClassFilter ||
    hasSnapshotFilter ||
    hasRegionFilter ||
    hasLabelsFilter ||
    hasLastEventFilter ||
    hasResourceFilter

  const handleChangeFilter = (value: string) => {
    table.setGlobalFilter(value)
    table.setPageIndex(0)
  }

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center gap-2">
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
              <Button variant="outline" className="shrink-0" aria-label="Filter">
                <ListFilter className="w-4 h-4" />
                <span className="max-[420px]:hidden">Filter</span>
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
                  <Boxes className="w-4 h-4" />
                  Class
                </DropdownMenuSubTrigger>
                <DropdownMenuPortal>
                  <DropdownMenuSubContent className="p-0 w-64">
                    <SandboxClassFilter
                      value={(table.getColumn('class')?.getFilterValue() as string[]) || []}
                      onFilterChange={(value) => table.getColumn('class')?.setFilterValue(value)}
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

        <div className="flex shrink-0 items-center gap-2 sm:ml-auto">
          <SandboxTableSettings table={table} />
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

          {hasClassFilter && (
            <SandboxClassFilterIndicator
              value={(table.getColumn('class')?.getFilterValue() as string[]) || []}
              onFilterChange={(value) => table.getColumn('class')?.setFilterValue(value)}
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

function SandboxTableSettings({ table }: Pick<SandboxTableHeaderProps, 'table'>) {
  const hideableColumns = table.getAllLeafColumns().filter((column) => column.getCanHide())

  if (hideableColumns.length === 0) {
    return null
  }

  return (
    <DropdownMenu modal={false}>
      <Tooltip>
        <TooltipTrigger asChild>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="icon-sm" aria-label="Table settings">
              <Columns className="size-4" />
            </Button>
          </DropdownMenuTrigger>
        </TooltipTrigger>
        <TooltipContent>Table settings</TooltipContent>
      </Tooltip>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuLabel>Columns</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {hideableColumns.map((column) => (
          <DropdownMenuCheckboxItem
            key={column.id}
            checked={column.getIsVisible()}
            onCheckedChange={(checked) => column.toggleVisibility(checked === true)}
            onSelect={(event) => event.preventDefault()}
          >
            {SANDBOX_TABLE_COLUMN_LABELS[column.id] ?? column.id}
          </DropdownMenuCheckboxItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
