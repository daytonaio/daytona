/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Badge } from '@/components/ui/badge'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { InputGroup, InputGroupAddon } from '@/components/ui/input-group'
import { Popover, PopoverAnchor, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Spinner } from '@/components/ui/spinner'
import { useRegionLookup } from '@/hooks/queries/useRegionsQuery'
import { useSnapshotsQuery } from '@/hooks/queries/useSnapshotsQuery'
import { useDebouncedValue } from '@/hooks/useDebouncedValue'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import { SnapshotState, type SnapshotDto } from '@daytona/api-client'
import { AnimatePresence, motion } from 'framer-motion'
import { Check, ChevronDownIcon, SearchIcon, X } from 'lucide-react'
import { MouseEvent, Ref, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { Tooltip } from '../Tooltip'

const EMPTY_SNAPSHOTS: SnapshotDto[] = []

const commandInputIconMotion = {
  initial: { opacity: 0, y: 4, scale: 0.95 },
  animate: { opacity: 1, y: 0, scale: 1 },
  exit: { opacity: 0, y: -4, scale: 0.95 },
  transition: { duration: 0.14, ease: 'easeOut' },
} as const

function CommandInputStatusIcon({ fetching }: { fetching: boolean }) {
  return (
    <span className="relative flex size-4 shrink-0 items-center justify-center opacity-50">
      <AnimatePresence initial={false} mode="wait">
        {fetching ? (
          <motion.span
            key="fetching"
            className="absolute inset-0 flex items-center justify-center"
            {...commandInputIconMotion}
          >
            <Spinner className="size-4" />
          </motion.span>
        ) : (
          <motion.span
            key="search"
            className="absolute inset-0 flex items-center justify-center"
            {...commandInputIconMotion}
          >
            <SearchIcon className="size-4" />
          </motion.span>
        )}
      </AnimatePresence>
    </span>
  )
}

function SnapshotRegionInline({
  regionIds,
  getRegionName,
}: {
  regionIds?: string[]
  getRegionName: (regionId: string) => string | undefined
}) {
  if (!regionIds?.length) {
    return <span className="ml-auto shrink-0 text-xs text-muted-foreground">-</span>
  }

  const regionNames = regionIds.map((regionId) => getRegionName(regionId) ?? regionId)
  const firstRegion = regionNames[0]
  const remainingCount = regionNames.length - 1

  if (remainingCount === 0) {
    return (
      <span
        className="ml-auto block max-w-[150px] shrink-0 truncate text-right text-xs text-muted-foreground"
        title={firstRegion}
      >
        {firstRegion}
      </span>
    )
  }

  return (
    <Tooltip
      label={
        <div className="ml-auto flex shrink-0 items-center gap-1.5">
          <span className="max-w-[150px] truncate text-xs text-muted-foreground">{firstRegion}</span>
          <Badge variant="secondary" className="h-5 px-1.5 py-0 text-xs">
            +{remainingCount}
          </Badge>
        </div>
      }
      content={
        <div className="flex flex-col gap-1">
          {regionNames.map((regionName, index) => (
            <span key={`${regionName}-${index}`}>{regionName}</span>
          ))}
        </div>
      }
    />
  )
}

function getSnapshotSearchValue(snapshot: SnapshotDto, getRegionName: (regionId: string) => string | undefined) {
  const regions = (snapshot.regionIds ?? []).map((regionId) => getRegionName(regionId) ?? regionId).join(' ')
  return `${snapshot.name} ${regions}`
}

export interface SnapshotSelectProps {
  ref?: Ref<SnapshotSelectRef>
  id?: string
  name?: string
  value?: string
  disabled?: boolean
  placeholder?: string
  pageSize?: number
  className?: string
  popoverContainer?: HTMLElement | null
  onOpenChange?: (open: boolean) => void
  onSnapshotChange?: (snapshot: SnapshotDto) => void
  onValueChange?: (value: string | undefined) => void
}

export interface SnapshotSelectRef {
  open: () => void
}

export function SnapshotSelect({
  ref,
  id,
  name,
  value,
  disabled,
  placeholder = 'Select snapshot',
  pageSize = 100,
  className,
  popoverContainer,
  onOpenChange,
  onSnapshotChange,
  onValueChange,
}: SnapshotSelectProps) {
  const [open, setOpen] = useState(false)
  const [searchValue, setSearchValue] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const debouncedSearchValue = useDebouncedValue(searchValue, 300)
  const searchTerm = debouncedSearchValue.trim()
  const { selectedOrganization } = useSelectedOrganization()
  const { getRegionName } = useRegionLookup(selectedOrganization?.id)

  const {
    data: snapshotsData,
    isLoading: snapshotsLoading,
    isFetching: snapshotsFetching,
  } = useSnapshotsQuery(
    {
      page: 1,
      pageSize,
    },
    { enabled: !disabled },
  )

  const canSearchServer = Boolean(snapshotsData && snapshotsData.total > pageSize)
  const shouldFetchSearchResults = open && canSearchServer && Boolean(searchTerm)
  const { data: searchedSnapshotsData, isFetching: searchedSnapshotsFetching } = useSnapshotsQuery(
    {
      page: 1,
      pageSize,
      filters: { name: searchTerm },
    },
    { enabled: !disabled && shouldFetchSearchResults },
  )

  const sourceSnapshots = shouldFetchSearchResults ? searchedSnapshotsData?.items : snapshotsData?.items
  const snapshots = useMemo(
    () => sourceSnapshots?.filter((snapshot) => snapshot.state === SnapshotState.ACTIVE) ?? EMPTY_SNAPSHOTS,
    [sourceSnapshots],
  )
  const selectedSnapshot = useMemo(() => snapshots.find((snapshot) => snapshot.name === value), [snapshots, value])

  const loading = snapshotsLoading && !snapshotsData
  const searchPending = open && canSearchServer && searchValue.trim() !== searchTerm
  const fetching =
    searchPending || (shouldFetchSearchResults ? searchedSnapshotsFetching : snapshotsFetching && !loading)
  const selectedLabel = selectedSnapshot?.name ?? value ?? placeholder

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      onOpenChange?.(nextOpen)
      setOpen(nextOpen)
      if (nextOpen) {
        setSearchValue('')
      }
    },
    [onOpenChange],
  )

  useEffect(() => {
    if (open) {
      inputRef.current?.focus()
    }
  }, [open])

  useImperativeHandle(
    ref,
    () => ({
      open: () => handleOpenChange(true),
    }),
    [handleOpenChange],
  )

  const handleChange = (snapshot: SnapshotDto) => {
    onValueChange?.(snapshot.name)
    onSnapshotChange?.(snapshot)
    setSearchValue(snapshot.name)
    handleOpenChange(false)
  }

  const handleClear = (event: MouseEvent<HTMLButtonElement>) => {
    event.preventDefault()
    event.stopPropagation()

    onValueChange?.(undefined)
    setSearchValue('')
    handleOpenChange(false)
  }

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      {name && <input type="hidden" name={name} value={value ?? ''} />}
      <PopoverAnchor asChild>
        <InputGroup
          className={cn(
            'h-8 overflow-hidden data-[disabled]:opacity-50',
            {
              'opacity-50': loading,
            },
            className,
          )}
          data-disabled={loading || disabled ? true : undefined}
        >
          <PopoverTrigger asChild>
            <button
              id={id}
              type="button"
              disabled={loading || disabled}
              data-slot="input-group-control"
              className="absolute inset-0 z-10 rounded-md outline-none disabled:cursor-not-allowed"
              aria-label={loading ? 'Loading snapshots' : selectedLabel}
            />
          </PopoverTrigger>
          <span
            aria-hidden="true"
            className={cn('min-w-0 flex-1 truncate px-3 text-sm', {
              'text-muted-foreground': !value,
            })}
          >
            {loading ? 'Loading snapshots...' : selectedLabel}
          </span>
          <InputGroupAddon align="inline-end" className="text-foreground">
            <span className="flex size-5 items-center justify-center">
              {value && (
                <button
                  type="button"
                  aria-label="Clear snapshot"
                  disabled={loading || disabled}
                  className="relative z-20 rounded-sm p-0.5 text-current opacity-50 hover:text-foreground disabled:cursor-not-allowed"
                  onClick={handleClear}
                >
                  <X aria-hidden="true" className="size-3.5" />
                </button>
              )}
            </span>
            <ChevronDownIcon aria-hidden="true" className="size-4 text-current opacity-50" />
          </InputGroupAddon>
        </InputGroup>
      </PopoverAnchor>
      <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start" container={popoverContainer}>
        <Command shouldFilter={!shouldFetchSearchResults}>
          <CommandInput
            ref={inputRef}
            value={searchValue}
            onValueChange={setSearchValue}
            placeholder="Search snapshots..."
            icon={<CommandInputStatusIcon fetching={fetching} />}
          />
          <CommandList className="max-h-64 overscroll-contain">
            {loading ? (
              <div className="flex items-center justify-center gap-2 py-6 text-sm text-muted-foreground">
                <Spinner className="size-4" />
                Loading snapshots...
              </div>
            ) : (
              <>
                <CommandEmpty>{fetching ? 'Searching snapshots...' : 'No active snapshots found.'}</CommandEmpty>
                <CommandGroup
                  className={cn('transition-opacity duration-150 ease-out', {
                    'opacity-50': fetching && snapshots.length > 0,
                  })}
                >
                  {snapshots.map((snapshot) => (
                    <CommandItem
                      key={snapshot.id}
                      value={getSnapshotSearchValue(snapshot, getRegionName)}
                      onSelect={() => handleChange(snapshot)}
                      className="cursor-pointer gap-2"
                    >
                      <Check
                        className={cn('size-4 shrink-0', {
                          'opacity-100': value === snapshot.name,
                          'opacity-0': value !== snapshot.name,
                        })}
                      />
                      <div className="flex min-w-0 flex-1 items-center gap-2">
                        <span className="truncate">{snapshot.name}</span>
                        {snapshot.general && (
                          <Badge variant="secondary" className="shrink-0">
                            System
                          </Badge>
                        )}
                      </div>
                      <SnapshotRegionInline regionIds={snapshot.regionIds} getRegionName={getRegionName} />
                    </CommandItem>
                  ))}
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
