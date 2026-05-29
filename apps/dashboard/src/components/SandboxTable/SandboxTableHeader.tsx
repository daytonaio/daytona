/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Boxes,
  Calendar,
  CalendarPlus,
  Camera,
  Columns,
  Cpu,
  Eye,
  Globe,
  HardDrive,
  ListFilter,
  MemoryStick,
  RefreshCw,
  Square,
  Tag,
  Wrench,
} from 'lucide-react'
import { SearchInput } from '../SearchInput'
import TooltipButton from '../TooltipButton'
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
import { BooleanFilter, BooleanFilterIndicator } from './filters/BooleanFilter'
import { CreatedAtFilter, CreatedAtFilterIndicator } from './filters/CreatedAtFilter'
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
  regionOptions,
  regionsDataIsLoading,
  snapshots,
  snapshotsDataIsLoading,
  snapshotsDataHasMore,
  onChangeSnapshotSearchValue,
  onRefresh,
  isRefreshing = false,
}: SandboxTableHeaderProps) {
  const classColumnAvailable = Boolean(table.getColumn('sandboxClass'))
  const hasStateFilter = ((table.getColumn('state')?.getFilterValue() as string[]) || []).length > 0
  const hasClassFilter = ((table.getColumn('sandboxClass')?.getFilterValue() as string[]) || []).length > 0
  const hasSnapshotFilter = ((table.getColumn('snapshot')?.getFilterValue() as string[]) || []).length > 0
  const hasRegionFilter = ((table.getColumn('region')?.getFilterValue() as string[]) || []).length > 0
  const hasLabelsFilter = ((table.getColumn('labels')?.getFilterValue() as string[]) || []).length > 0
  const hasLastEventFilter = ((table.getColumn('lastEvent')?.getFilterValue() as Date[]) || []).length > 0
  const hasCreatedAtFilter = ((table.getColumn('createdAt')?.getFilterValue() as Date[]) || []).length > 0
  const hasIsPublicFilter = table.getColumn('isPublic')?.getFilterValue() !== undefined
  const hasIsRecoverableFilter = table.getColumn('isRecoverable')?.getFilterValue() !== undefined
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
    hasCreatedAtFilter ||
    hasIsPublicFilter ||
    hasIsRecoverableFilter ||
    hasResourceFilter

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center gap-2">
        <div className="flex flex-1 items-center gap-2 min-w-0">
          <SearchInput
            debounced
            value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
            onValueChange={(value) => table.getColumn('name')?.setFilterValue(value)}
            placeholder="Search by Name"
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
              {classColumnAvailable && (
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger>
                    <Boxes className="w-4 h-4" />
                    Class
                  </DropdownMenuSubTrigger>
                  <DropdownMenuPortal>
                    <DropdownMenuSubContent className="p-0 w-64">
                      <SandboxClassFilter
                        value={(table.getColumn('sandboxClass')?.getFilterValue() as string[]) || []}
                        onFilterChange={(value) => table.getColumn('sandboxClass')?.setFilterValue(value)}
                      />
                    </DropdownMenuSubContent>
                  </DropdownMenuPortal>
                </DropdownMenuSub>
              )}
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
                      snapshotsDataIsLoading={snapshotsDataIsLoading}
                      snapshotsDataHasMore={snapshotsDataHasMore}
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
              <DropdownMenuSub>
                <DropdownMenuSubTrigger>
                  <CalendarPlus className="w-4 h-4" />
                  Created
                </DropdownMenuSubTrigger>
                <DropdownMenuPortal>
                  <DropdownMenuSubContent className="p-3 w-92">
                    <CreatedAtFilter
                      onFilterChange={(value) => table.getColumn('createdAt')?.setFilterValue(value)}
                      value={(table.getColumn('createdAt')?.getFilterValue() as Date[]) || []}
                    />
                  </DropdownMenuSubContent>
                </DropdownMenuPortal>
              </DropdownMenuSub>
              <DropdownMenuSub>
                <DropdownMenuSubTrigger>
                  <Eye className="w-4 h-4" />
                  Public
                </DropdownMenuSubTrigger>
                <DropdownMenuPortal>
                  <DropdownMenuSubContent className="p-3 w-48">
                    <BooleanFilter
                      label="Public"
                      onFilterChange={(value) => table.getColumn('isPublic')?.setFilterValue(value)}
                      value={table.getColumn('isPublic')?.getFilterValue() as boolean | undefined}
                    />
                  </DropdownMenuSubContent>
                </DropdownMenuPortal>
              </DropdownMenuSub>
              <DropdownMenuSub>
                <DropdownMenuSubTrigger>
                  <Wrench className="w-4 h-4" />
                  Recoverable
                </DropdownMenuSubTrigger>
                <DropdownMenuPortal>
                  <DropdownMenuSubContent className="p-3 w-48">
                    <BooleanFilter
                      label="Recoverable"
                      onFilterChange={(value) => table.getColumn('isRecoverable')?.setFilterValue(value)}
                      value={table.getColumn('isRecoverable')?.getFilterValue() as boolean | undefined}
                    />
                  </DropdownMenuSubContent>
                </DropdownMenuPortal>
              </DropdownMenuSub>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <div className="flex shrink-0 items-center gap-2 sm:ml-auto">
          <TooltipButton
            variant="outline"
            size="icon-sm"
            onClick={onRefresh}
            disabled={isRefreshing}
            className="shrink-0"
            tooltipText="Refresh"
          >
            <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          </TooltipButton>
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
          {hasSnapshotFilter && (
            <SnapshotFilterIndicator
              value={(table.getColumn('snapshot')?.getFilterValue() as string[]) || []}
              onFilterChange={(value) => table.getColumn('snapshot')?.setFilterValue(value)}
              snapshots={snapshots}
              snapshotsDataIsLoading={snapshotsDataIsLoading}
              snapshotsDataHasMore={snapshotsDataHasMore}
              onChangeSnapshotSearchValue={onChangeSnapshotSearchValue}
            />
          )}
          {hasClassFilter && (
            <SandboxClassFilterIndicator
              value={(table.getColumn('sandboxClass')?.getFilterValue() as string[]) || []}
              onFilterChange={(value) => table.getColumn('sandboxClass')?.setFilterValue(value)}
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
            />
          )}
          {hasLastEventFilter && (
            <LastEventFilterIndicator
              value={(table.getColumn('lastEvent')?.getFilterValue() as Date[]) || []}
              onFilterChange={(value) => table.getColumn('lastEvent')?.setFilterValue(value)}
            />
          )}
          {hasCreatedAtFilter && (
            <CreatedAtFilterIndicator
              value={(table.getColumn('createdAt')?.getFilterValue() as Date[]) || []}
              onFilterChange={(value) => table.getColumn('createdAt')?.setFilterValue(value)}
            />
          )}
          {hasIsPublicFilter && (
            <BooleanFilterIndicator
              label="Public"
              value={table.getColumn('isPublic')?.getFilterValue() as boolean | undefined}
              onFilterChange={(value) => table.getColumn('isPublic')?.setFilterValue(value)}
            />
          )}
          {hasIsRecoverableFilter && (
            <BooleanFilterIndicator
              label="Recoverable"
              value={table.getColumn('isRecoverable')?.getFilterValue() as boolean | undefined}
              onFilterChange={(value) => table.getColumn('isRecoverable')?.setFilterValue(value)}
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
