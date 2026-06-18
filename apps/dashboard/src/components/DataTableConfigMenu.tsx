/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Column, ColumnPinningState, Table, VisibilityState } from '@tanstack/react-table'
import isEqual from 'fast-deep-equal'
import { Eye, EyeOff, GripVertical, Pin, PinOff, RotateCcw, Settings2Icon } from 'lucide-react'
import { motion, Reorder, useDragControls } from 'motion/react'
import { type ComponentProps, useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { useDeepCompareMemo } from '@/hooks/useDeepCompareMemo'
import { useStorageState } from '@/hooks/useStorageState'
import { cn } from '@/lib/utils'
import { Button } from './ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

const DATA_TABLE_CONFIG_STORAGE_PREFIX = 'daytona:data-table-config'
const DATA_TABLE_CONFIG_VERSION = 2
const DEFAULT_DATA_TABLE_CONFIG_EXCLUDED_COLUMN_IDS = ['actions', 'select'] as const
const DATA_TABLE_CONFIG_PIN_SEPARATOR_ID = '__data-table-config-pin-separator__'

type DataTableConfig = {
  columnOrder: string[]
  columnPinning: ColumnPinningState
  columnVisibility: VisibilityState
}

type DataTableConfigMenuProps<TData> = {
  align?: ComponentProps<typeof DropdownMenuContent>['align']
  className?: string
  excludedColumnIds?: readonly string[]
  getColumnLabel?: (columnId: string) => string
  persistenceKey: string
  table: Table<TData>
}

function DataTableConfigMenu<TData>({
  align = 'end',
  className,
  excludedColumnIds = DEFAULT_DATA_TABLE_CONFIG_EXCLUDED_COLUMN_IDS,
  getColumnLabel,
  persistenceKey,
  table,
}: DataTableConfigMenuProps<TData>) {
  const storageKey = useMemo(() => getTableConfigStorageKey(persistenceKey), [persistenceKey])
  const initialConfigRef = useRef<DataTableConfig | null>(null)
  const initialConfigStorageKeyRef = useRef<string | null>(null)
  const appliedStoredConfigKeyRef = useRef<string | null>(null)

  const [storedConfig, setStoredConfig, removeStoredConfig, { hasStoredValue: hasStoredConfig }] =
    useStorageState<DataTableConfig | null>(storageKey, null, {
      deserialize: deserializeTableConfig,
    })

  if (initialConfigStorageKeyRef.current !== storageKey) {
    initialConfigRef.current = getCurrentTableConfig(table)
    initialConfigStorageKeyRef.current = storageKey
  }

  const excludedColumnIdSet = useMemo(() => new Set(excludedColumnIds), [excludedColumnIds])
  const columns = table.getAllLeafColumns()
  const configurableColumns = columns.filter((column) => getColumnCanBeConfigured(column, excludedColumnIdSet))
  const configurableColumnIds = configurableColumns.map((column) => column.id)
  const configurableColumnsById = new Map(configurableColumns.map((column) => [column.id, column]))
  const currentGroupedColumnOrder = useDeepCompareMemo(
    getGroupedColumnOrder(columns, table.getState().columnOrder, table.getState().columnPinning, configurableColumnIds),
  )
  const [localGroupedColumnOrder, setLocalGroupedColumnOrder] = useState(currentGroupedColumnOrder)
  const localGroupedColumnOrderRef = useRef(localGroupedColumnOrder)
  const localGroupedColumnState = splitGroupedColumnOrder(localGroupedColumnOrder, configurableColumnIds)
  const hasPinnedConfigurableColumns = localGroupedColumnState.pinnedColumnIds.length > 0
  const currentConfig = getCurrentTableConfig(table)
  const initialConfig = initialConfigRef.current
  const hasConfigChanges = Boolean(initialConfig && !isEqual(currentConfig, initialConfig))

  useEffect(() => {
    setLocalGroupedColumnOrder(currentGroupedColumnOrder)
    localGroupedColumnOrderRef.current = currentGroupedColumnOrder
  }, [currentGroupedColumnOrder])

  const persistConfig = useCallback(
    (updates: Partial<DataTableConfig>) => {
      const nextConfig = normalizeTableConfig(
        {
          ...getCurrentTableConfig(table),
          ...updates,
        },
        table.getAllLeafColumns(),
      )

      if (initialConfigRef.current && isEqual(nextConfig, initialConfigRef.current)) {
        removeStoredConfig()
        return
      }

      setStoredConfig(nextConfig)
    },
    [removeStoredConfig, setStoredConfig, table],
  )

  const applyConfig = useCallback(
    (config: DataTableConfig) => {
      const normalizedConfig = normalizeTableConfig(config, table.getAllLeafColumns())

      table.setColumnOrder(normalizedConfig.columnOrder)
      table.setColumnVisibility(normalizedConfig.columnVisibility)
      table.setColumnPinning(normalizedConfig.columnPinning)
    },
    [table],
  )

  useEffect(() => {
    if (appliedStoredConfigKeyRef.current === storageKey) {
      return
    }

    appliedStoredConfigKeyRef.current = storageKey

    if (storedConfig) {
      applyConfig(storedConfig)
    }
  }, [applyConfig, storageKey, storedConfig])

  const handleVisibilityToggle = useCallback(
    (column: Column<TData, unknown>) => {
      const nextVisibility = {
        ...table.getState().columnVisibility,
        [column.id]: !column.getIsVisible(),
      }

      table.setColumnVisibility(nextVisibility)
      persistConfig({ columnVisibility: nextVisibility })
    },
    [persistConfig, table],
  )

  const handlePinToggle = useCallback(
    (column: Column<TData, unknown>) => {
      const currentPinning = normalizeColumnPinning(table.getState().columnPinning, table.getAllLeafColumns())
      const nextPinning = getNextColumnPinning(
        currentPinning,
        column.id,
        column.getIsPinned() === 'left' ? false : 'left',
      )

      table.setColumnPinning(nextPinning)
      persistConfig({ columnPinning: nextPinning })
    },
    [persistConfig, table],
  )

  const commitGroupedColumnOrder = useCallback(
    (nextGroupedColumnOrder: string[]) => {
      const nextConfig = getTableConfigUpdatesFromGroupedOrder(table, nextGroupedColumnOrder, configurableColumnIds)

      table.setColumnOrder(nextConfig.columnOrder)
      table.setColumnPinning(nextConfig.columnPinning)
      persistConfig(nextConfig)
    },
    [configurableColumnIds, persistConfig, table],
  )

  const handleLocalReorder = useCallback((nextGroupedColumnOrder: string[]) => {
    localGroupedColumnOrderRef.current = nextGroupedColumnOrder
    setLocalGroupedColumnOrder(nextGroupedColumnOrder)
  }, [])

  const handleDragEnd = useCallback(() => {
    commitGroupedColumnOrder(localGroupedColumnOrderRef.current)
  }, [commitGroupedColumnOrder])

  const handleUnpinAll = useCallback(() => {
    const groupedOrder = splitGroupedColumnOrder(localGroupedColumnOrderRef.current, configurableColumnIds)
    const nextGroupedColumnOrder = [DATA_TABLE_CONFIG_PIN_SEPARATOR_ID, ...groupedOrder.orderedColumnIds]

    localGroupedColumnOrderRef.current = nextGroupedColumnOrder
    setLocalGroupedColumnOrder(nextGroupedColumnOrder)
    commitGroupedColumnOrder(nextGroupedColumnOrder)
  }, [commitGroupedColumnOrder, configurableColumnIds])

  const handleReset = useCallback(() => {
    removeStoredConfig()

    const defaultConfig = initialConfigRef.current ?? getCurrentTableConfig(table)
    applyConfig(defaultConfig)
  }, [applyConfig, removeStoredConfig, table])

  if (configurableColumns.length === 0) {
    return null
  }

  return (
    <DropdownMenu modal={false}>
      <Tooltip>
        <TooltipTrigger asChild>
          <DropdownMenuTrigger asChild>
            <Button
              variant="outline"
              size="icon-sm"
              className={cn('relative shrink-0', className)}
              aria-label="Table settings"
            >
              <Settings2Icon className="size-4" />
              {hasStoredConfig && (
                <span
                  aria-hidden="true"
                  className="absolute right-1 top-1 size-1.5 rounded-full bg-info ring-2 ring-background"
                />
              )}
            </Button>
          </DropdownMenuTrigger>
        </TooltipTrigger>
        <TooltipContent>Table settings</TooltipContent>
      </Tooltip>
      <DropdownMenuContent align={align} className="w-80 p-1.5">
        <div className="flex items-center justify-between gap-2 px-1 py-1">
          <DropdownMenuLabel className="px-1 py-0">Columns</DropdownMenuLabel>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="h-7 px-2"
            disabled={!hasConfigChanges}
            onClick={handleReset}
          >
            <RotateCcw className="size-3.5" />
            Reset
          </Button>
        </div>
        <DropdownMenuSeparator />
        {hasPinnedConfigurableColumns && (
          <div className="flex items-center justify-between gap-2 px-2 pb-1 pt-1.5">
            <div className="text-xs font-medium text-muted-foreground">Pinned</div>
            <Button type="button" variant="ghost" size="sm" className="h-7 px-2" onClick={handleUnpinAll}>
              <PinOff className="size-3.5" />
              Unpin all
            </Button>
          </div>
        )}
        <motion.div layoutScroll className="scrollbar-sm scroll-fade max-h-80 overflow-y-auto pb-1">
          <Reorder.Group as="div" axis="y" values={localGroupedColumnOrder} onReorder={handleLocalReorder}>
            {localGroupedColumnOrder.map((columnId) => {
              if (columnId === DATA_TABLE_CONFIG_PIN_SEPARATOR_ID) {
                return <DataTableConfigMenuGroupSeparator key={DATA_TABLE_CONFIG_PIN_SEPARATOR_ID} />
              }

              const column = configurableColumnsById.get(columnId)
              if (!column) {
                return null
              }

              return (
                <DataTableConfigMenuItem
                  key={column.id}
                  column={column}
                  label={getColumnLabel?.(column.id) ?? getDefaultColumnLabel(column)}
                  onDragEnd={handleDragEnd}
                  onPinToggle={handlePinToggle}
                  onVisibilityToggle={handleVisibilityToggle}
                />
              )
            })}
          </Reorder.Group>
        </motion.div>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

type DataTableConfigMenuItemProps<TData> = {
  column: Column<TData, unknown>
  label: string
  onDragEnd: () => void
  onPinToggle: (column: Column<TData, unknown>) => void
  onVisibilityToggle: (column: Column<TData, unknown>) => void
}

function DataTableConfigMenuItem<TData>({
  column,
  label,
  onDragEnd,
  onPinToggle,
  onVisibilityToggle,
}: DataTableConfigMenuItemProps<TData>) {
  const dragControls = useDragControls()
  const isPinned = column.getIsPinned() === 'left'

  return (
    <Reorder.Item
      as="div"
      value={column.id}
      dragControls={dragControls}
      dragListener={false}
      onDragEnd={onDragEnd}
      transition={{ type: 'spring', bounce: 0, duration: 0.12 }}
      className="relative grid grid-cols-[1.75rem_minmax(0,1fr)_auto] items-center gap-2 rounded-md bg-popover px-1.5 py-1 text-sm hover:bg-accent/60"
    >
      <button
        type="button"
        aria-label={`Reorder ${label}`}
        className="flex size-7 cursor-grab touch-none items-center justify-center rounded-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground active:cursor-grabbing"
        onPointerDown={(event) => {
          if (event.button !== 0) {
            return
          }

          dragControls.start(event)
        }}
      >
        <GripVertical className="size-4" />
      </button>

      <span
        className={cn('min-w-0 truncate', {
          'text-muted-foreground': !column.getIsVisible(),
        })}
      >
        {label}
      </span>

      <div className="flex items-center gap-0.5">
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          className="size-7 text-muted-foreground hover:text-foreground"
          disabled={!column.getCanHide()}
          aria-label={`${column.getIsVisible() ? 'Hide' : 'Show'} ${label}`}
          onClick={(event) => {
            event.preventDefault()
            event.stopPropagation()
            onVisibilityToggle(column)
          }}
        >
          {column.getIsVisible() ? <Eye className="size-4" /> : <EyeOff className="size-4" />}
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          className={cn('size-7 text-muted-foreground hover:text-foreground', {
            'text-foreground': isPinned,
          })}
          disabled={!column.getCanPin()}
          aria-label={`${isPinned ? 'Unpin' : 'Pin'} ${label}`}
          onClick={(event) => {
            event.preventDefault()
            event.stopPropagation()
            onPinToggle(column)
          }}
        >
          {isPinned ? <PinOff className="size-4" /> : <Pin className="size-4" />}
        </Button>
      </div>
    </Reorder.Item>
  )
}

function DataTableConfigMenuGroupSeparator() {
  return (
    <Reorder.Item
      as="div"
      value={DATA_TABLE_CONFIG_PIN_SEPARATOR_ID}
      dragListener={false}
      transition={{ type: 'spring', bounce: 0, duration: 0.12 }}
      className="cursor-default px-2 pb-1 pt-2 text-xs font-medium text-muted-foreground"
    >
      Unpinned
    </Reorder.Item>
  )
}

function getTableConfigStorageKey(persistenceKey: string) {
  return `${DATA_TABLE_CONFIG_STORAGE_PREFIX}:${persistenceKey}:v${DATA_TABLE_CONFIG_VERSION}`
}

function getDefaultColumnLabel<TData>(column: Column<TData, unknown>) {
  return typeof column.columnDef.header === 'string' ? column.columnDef.header : column.id
}

function getColumnCanBeConfigured<TData>(column: Column<TData, unknown>, excludedColumnIds: Set<string>) {
  if (excludedColumnIds.has(column.id)) {
    return false
  }

  return Boolean(column.columnDef.header || column.columnDef.cell || column.getCanHide() || column.getIsPinned())
}

function getGroupedColumnOrder<TData>(
  columns: Column<TData, unknown>[],
  columnOrder: unknown,
  columnPinning: unknown,
  configurableColumnIds: string[],
) {
  const configurableColumnIdSet = new Set(configurableColumnIds)
  const normalizedColumnOrder = normalizeColumnOrder(
    columnOrder,
    columns.map((column) => column.id),
  ).filter((columnId) => configurableColumnIdSet.has(columnId))
  const normalizedColumnPinning = normalizeColumnPinning(columnPinning, columns)
  const pinnedColumnIds = new Set(normalizedColumnPinning.left ?? [])

  return [
    ...normalizedColumnOrder.filter((columnId) => pinnedColumnIds.has(columnId)),
    DATA_TABLE_CONFIG_PIN_SEPARATOR_ID,
    ...normalizedColumnOrder.filter((columnId) => !pinnedColumnIds.has(columnId)),
  ]
}

function getCurrentTableConfig<TData>(table: Table<TData>): DataTableConfig {
  const columns = table.getAllLeafColumns()
  const columnIds = columns.map((column) => column.id)

  return normalizeTableConfig(
    {
      columnOrder: normalizeColumnOrder(table.getState().columnOrder, columnIds),
      columnPinning: table.getState().columnPinning,
      columnVisibility: table.getState().columnVisibility,
    },
    columns,
  )
}

function normalizeTableConfig<TData>(config: DataTableConfig, columns: Column<TData, unknown>[]): DataTableConfig {
  const columnIds = columns.map((column) => column.id)
  const columnOrder = normalizeColumnOrder(config.columnOrder, columnIds)

  return {
    columnOrder,
    columnPinning: orderColumnPinning(normalizeColumnPinning(config.columnPinning, columns), columnOrder),
    columnVisibility: normalizeColumnVisibility(config.columnVisibility, columns),
  }
}

function normalizeColumnOrder(columnOrder: unknown, columnIds: string[]) {
  const columnIdSet = new Set(columnIds)
  const orderedIds = Array.isArray(columnOrder)
    ? columnOrder.filter((columnId): columnId is string => typeof columnId === 'string' && columnIdSet.has(columnId))
    : []
  const orderedIdSet = new Set(orderedIds)
  const remainingIds = columnIds.filter((columnId) => !orderedIdSet.has(columnId))

  return insertMissingColumnIds(orderedIds, remainingIds, columnIds)
}

function insertMissingColumnIds(orderedIds: string[], missingIds: string[], columnIds: string[]) {
  const nextOrder = [...orderedIds]

  for (const missingId of missingIds) {
    const defaultIndex = columnIds.indexOf(missingId)
    const previousColumnId = columnIds
      .slice(0, defaultIndex)
      .reverse()
      .find((columnId) => nextOrder.includes(columnId))

    if (previousColumnId) {
      nextOrder.splice(nextOrder.indexOf(previousColumnId) + 1, 0, missingId)
      continue
    }

    const nextColumnId = columnIds.slice(defaultIndex + 1).find((columnId) => nextOrder.includes(columnId))

    if (nextColumnId) {
      nextOrder.splice(nextOrder.indexOf(nextColumnId), 0, missingId)
      continue
    }

    nextOrder.push(missingId)
  }

  return nextOrder
}

function normalizeColumnVisibility<TData>(columnVisibility: unknown, columns: Column<TData, unknown>[]) {
  const columnsById = new Map(columns.map((column) => [column.id, column]))
  const normalizedVisibility: VisibilityState = {}

  if (!columnVisibility || typeof columnVisibility !== 'object' || Array.isArray(columnVisibility)) {
    return normalizedVisibility
  }

  for (const [columnId, isVisible] of Object.entries(columnVisibility)) {
    const column = columnsById.get(columnId)
    if (!column?.getCanHide() || typeof isVisible !== 'boolean') {
      continue
    }

    normalizedVisibility[columnId] = isVisible
  }

  return normalizedVisibility
}

function normalizeColumnPinning<TData>(columnPinning: unknown, columns: Column<TData, unknown>[]) {
  const columnsById = new Map(columns.map((column) => [column.id, column]))
  const usedColumnIds = new Set<string>()

  if (!columnPinning || typeof columnPinning !== 'object' || Array.isArray(columnPinning)) {
    return {}
  }

  const pinning = columnPinning as ColumnPinningState
  const readPinnedIds = (side: 'left' | 'right') => {
    if (!Array.isArray(pinning[side])) {
      return []
    }

    return pinning[side].filter((columnId): columnId is string => {
      const column = columnsById.get(columnId)
      if (!column?.getCanPin() || usedColumnIds.has(columnId)) {
        return false
      }

      usedColumnIds.add(columnId)
      return true
    })
  }

  return {
    left: readPinnedIds('left'),
    right: readPinnedIds('right'),
  }
}

function getNextColumnPinning(
  columnPinning: ColumnPinningState,
  columnId: string,
  pinSide: 'left' | 'right' | false,
): ColumnPinningState {
  const nextPinning: ColumnPinningState = {
    left: (columnPinning.left ?? []).filter((pinnedColumnId) => pinnedColumnId !== columnId),
    right: (columnPinning.right ?? []).filter((pinnedColumnId) => pinnedColumnId !== columnId),
  }

  if (pinSide) {
    nextPinning[pinSide] = [...(nextPinning[pinSide] ?? []), columnId]
  }

  return nextPinning
}

function getTableConfigUpdatesFromGroupedOrder<TData>(
  table: Table<TData>,
  groupedColumnOrder: string[],
  configurableColumnIds: string[],
): Pick<DataTableConfig, 'columnOrder' | 'columnPinning'> {
  const columns = table.getAllLeafColumns()
  const columnIds = columns.map((column) => column.id)
  const groupedOrder = splitGroupedColumnOrder(groupedColumnOrder, configurableColumnIds)
  const currentOrder = normalizeColumnOrder(table.getState().columnOrder, columnIds)
  const nextOrder = mergeConfigurableColumnOrder(currentOrder, groupedOrder.orderedColumnIds, configurableColumnIds)
  const nextPinning = getColumnPinningFromGroupedOrder(
    normalizeColumnPinning(table.getState().columnPinning, columns),
    configurableColumnIds,
    groupedOrder.pinnedColumnIds,
    nextOrder,
  )

  return {
    columnOrder: nextOrder,
    columnPinning: nextPinning,
  }
}

function splitGroupedColumnOrder(groupedColumnOrder: string[], configurableColumnIds: string[]) {
  const normalizedGroupedColumnOrder = normalizeGroupedColumnOrder(groupedColumnOrder, configurableColumnIds)
  const separatorIndex = normalizedGroupedColumnOrder.indexOf(DATA_TABLE_CONFIG_PIN_SEPARATOR_ID)
  const pinnedColumnIds = normalizedGroupedColumnOrder.slice(0, separatorIndex)
  const unpinnedColumnIds = normalizedGroupedColumnOrder.slice(separatorIndex + 1)

  return {
    orderedColumnIds: [...pinnedColumnIds, ...unpinnedColumnIds],
    pinnedColumnIds,
    unpinnedColumnIds,
  }
}

function normalizeGroupedColumnOrder(groupedColumnOrder: unknown, configurableColumnIds: string[]) {
  const configurableColumnIdSet = new Set(configurableColumnIds)
  const seenColumnIds = new Set<string>()
  const normalizedGroupedColumnOrder: string[] = []
  let hasSeparator = false

  if (Array.isArray(groupedColumnOrder)) {
    for (const columnId of groupedColumnOrder) {
      if (columnId === DATA_TABLE_CONFIG_PIN_SEPARATOR_ID && !hasSeparator) {
        normalizedGroupedColumnOrder.push(columnId)
        hasSeparator = true
        continue
      }

      if (typeof columnId !== 'string' || !configurableColumnIdSet.has(columnId) || seenColumnIds.has(columnId)) {
        continue
      }

      normalizedGroupedColumnOrder.push(columnId)
      seenColumnIds.add(columnId)
    }
  }

  if (!hasSeparator) {
    normalizedGroupedColumnOrder.push(DATA_TABLE_CONFIG_PIN_SEPARATOR_ID)
  }

  for (const columnId of configurableColumnIds) {
    if (!seenColumnIds.has(columnId)) {
      normalizedGroupedColumnOrder.push(columnId)
    }
  }

  return normalizedGroupedColumnOrder
}

function mergeConfigurableColumnOrder(
  columnOrder: string[],
  nextConfigurableOrder: string[],
  configurableColumnIds: string[],
) {
  const configurableColumnIdSet = new Set(configurableColumnIds)
  const nextConfigurableIds = normalizeColumnOrder(nextConfigurableOrder, configurableColumnIds)
  const nextConfigurableIdQueue = [...nextConfigurableIds]

  return columnOrder.map((columnId) => {
    if (!configurableColumnIdSet.has(columnId)) {
      return columnId
    }

    return nextConfigurableIdQueue.shift() ?? columnId
  })
}

function getColumnPinningFromGroupedOrder(
  columnPinning: ColumnPinningState,
  configurableColumnIds: string[],
  pinnedColumnIds: string[],
  columnOrder: string[],
) {
  const configurableColumnIdSet = new Set(configurableColumnIds)

  return orderColumnPinning(
    {
      left: [
        ...(columnPinning.left ?? []).filter((columnId) => !configurableColumnIdSet.has(columnId)),
        ...pinnedColumnIds,
      ],
      right: (columnPinning.right ?? []).filter((columnId) => !configurableColumnIdSet.has(columnId)),
    },
    columnOrder,
  )
}

function orderColumnPinning(columnPinning: ColumnPinningState, columnOrder: string[]) {
  const columnOrderIndex = new Map(columnOrder.map((columnId, index) => [columnId, index]))
  const orderPinnedIds = (columnIds: string[] = []) =>
    columnIds
      .filter((columnId) => columnOrderIndex.has(columnId))
      .sort((a, b) => (columnOrderIndex.get(a) ?? 0) - (columnOrderIndex.get(b) ?? 0))

  return {
    left: orderPinnedIds(columnPinning.left),
    right: orderPinnedIds(columnPinning.right),
  }
}

function deserializeTableConfig(storedConfig: string): DataTableConfig {
  const parsedConfig = JSON.parse(storedConfig)

  if (!parsedConfig || typeof parsedConfig !== 'object') {
    throw new Error('Stored table configuration is invalid')
  }

  return parsedConfig as DataTableConfig
}

export { DataTableConfigMenu }
