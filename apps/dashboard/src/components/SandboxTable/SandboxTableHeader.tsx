/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ListFilter, Square, Globe, Camera, RefreshCw } from 'lucide-react'
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
import { SnapshotFilter, SnapshotFilterIndicator } from './filters/SnapshotFilter'

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
  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center mb-4">
      <div className="flex gap-2 items-center">
        <DebouncedInput
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(value) => table.getColumn('name')?.setFilterValue(value)}
          placeholder="Search by Name"
          className="max-w-[200px]"
        />

        <Button
          variant="outline"
          size="sm"
          onClick={onRefresh}
          disabled={isRefreshing}
          className="flex items-center gap-2"
        >
          <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          Refresh
        </Button>

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
      </div>
    </div>
  )
}
