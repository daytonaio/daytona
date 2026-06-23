/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationSuspendedError } from '@/api/errors'
import { useRegisterCommands, type CommandConfig } from '@/components/CommandPalette'
import { CursorPagination } from '@/components/CursorPagination'
import { ForkTreeDialog } from '@/components/ForkTreeDialog'
import { PageContent, PageFooter, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { RecursiveDeleteDialog } from '@/components/RecursiveDeleteDialog'
import { CreateSandboxSheet } from '@/components/Sandbox/CreateSandboxSheet'
import { CreateSshAccessSheet } from '@/components/sandboxes/CreateSshAccessSheet'
import { RevokeSshAccessDialog } from '@/components/sandboxes/RevokeSshAccessDialog'
import SandboxDetailsSheet, {
  type SandboxDetailsSheetTabValue,
  type SandboxSheetRef,
} from '@/components/sandboxes/SandboxDetailsSheet'
import { tabParser } from '@/components/sandboxes/SearchParams'
import { SandboxTable } from '@/components/SandboxTable'
import type { SandboxTableRef } from '@/components/SandboxTable/types'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from '@/constants/Pagination'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { RoutePath } from '@/enums/RoutePath'
import { mutationKeys } from '@/hooks/mutations/mutationKeys'
import { useArchiveSandboxMutation } from '@/hooks/mutations/useArchiveSandboxMutation'
import { useDeleteSandboxMutation } from '@/hooks/mutations/useDeleteSandboxMutation'
import { useMutatingSandboxes } from '@/hooks/mutations/useMutatingSandboxes'
import { useRecoverSandboxMutation } from '@/hooks/mutations/useRecoverSandboxMutation'
import { useStartSandboxMutation } from '@/hooks/mutations/useStartSandboxMutation'
import { useStopSandboxMutation } from '@/hooks/mutations/useStopSandboxMutation'
import { usePauseSandboxMutation } from '@/hooks/mutations/usePauseSandboxMutation'
import { queryKeys } from '@/hooks/queries/queryKeys'
import {
  DEFAULT_SANDBOX_SORTING,
  SandboxFilters,
  SandboxQueryParams,
  SandboxSorting,
  useSandboxesQuery,
} from '@/hooks/queries/useSandboxesQuery'
import { SnapshotFilters, SnapshotQueryParams, useSnapshotsQuery } from '@/hooks/queries/useSnapshotsQuery'
import { useAvailableRegionsQuery, useRegionLookup } from '@/hooks/queries/useRegionsQuery'
import { useApi } from '@/hooks/useApi'
import { useConfig } from '@/hooks/useConfig'
import { useSandboxWsSync, type SandboxWsSyncEvent } from '@/hooks/useSandboxWsSync'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { createBulkActionToast } from '@/lib/bulk-action-toast'
import { handleApiError } from '@/lib/error-handling'
import { getLocalStorageItem, setLocalStorageItem } from '@/lib/local-storage'
import { EMPTY_REGIONS } from '@/lib/regions'
import { formatDuration, pluralize } from '@/lib/utils'
import {
  ListSandboxesResponse,
  OrganizationRolePermissionsEnum,
  OrganizationUserRoleEnum,
  Sandbox,
  SandboxClass,
  SandboxDesiredState,
  SandboxListItem,
  SandboxListSortDirection,
  SandboxListSortField,
  SandboxState,
} from '@daytona/api-client'
import type { Sandbox as CreatedSandbox } from '@daytona/sdk'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { PlusIcon } from 'lucide-react'
import {
  parseAsArrayOf,
  parseAsBoolean,
  parseAsFloat,
  parseAsInteger,
  parseAsIsoDateTime,
  parseAsJson,
  parseAsString,
  useQueryState,
  useQueryStates,
} from 'nuqs'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'

type CreateSandboxSnapshotVariables = {
  sandboxId: string
  name: string
  includeMemory: boolean
}

const SANDBOX_SORT_FIELDS: SandboxListSortField[] = [
  SandboxListSortField.NAME,
  SandboxListSortField.LAST_ACTIVITY_AT,
  SandboxListSortField.CREATED_AT,
]
const SANDBOX_SORT_DIRECTIONS = Object.values(SandboxListSortDirection)
const SANDBOX_STATES = Object.values(SandboxState)
const SANDBOX_CLASSES = Object.values(SandboxClass)
const DEFAULT_SANDBOXES: SandboxListItem[] = []
const SANDBOX_LIST_REVALIDATION_DEBOUNCE_MS = 2000

const labelsParser = parseAsJson<Record<string, string>>((value) => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return null
  }

  const entries = Object.entries(value).filter(
    (entry): entry is [string, string] => typeof entry[0] === 'string' && typeof entry[1] === 'string',
  )

  return Object.fromEntries(entries)
}).withDefault({})

const sandboxViewSearchParams = {
  limit: parseAsInteger.withDefault(DEFAULT_PAGE_SIZE),
  search: parseAsString.withDefault(''),
  states: parseAsArrayOf(parseAsString).withDefault([]),
  snapshots: parseAsArrayOf(parseAsString).withDefault([]),
  regions: parseAsArrayOf(parseAsString).withDefault([]),
  sandboxClasses: parseAsArrayOf(parseAsString).withDefault([]),
  labels: labelsParser,
  minCpu: parseAsFloat,
  maxCpu: parseAsFloat,
  minMemory: parseAsFloat,
  maxMemory: parseAsFloat,
  minDisk: parseAsFloat,
  maxDisk: parseAsFloat,
  lastEventAfter: parseAsIsoDateTime,
  lastEventBefore: parseAsIsoDateTime,
  createdAtAfter: parseAsIsoDateTime,
  createdAtBefore: parseAsIsoDateTime,
  isPublic: parseAsBoolean,
  isRecoverable: parseAsBoolean,
  sort: parseAsString.withDefault(DEFAULT_SANDBOX_SORTING.field ?? SandboxListSortField.LAST_ACTIVITY_AT),
  order: parseAsString.withDefault(DEFAULT_SANDBOX_SORTING.direction ?? SandboxListSortDirection.DESC),
}

function normalizePageSize(pageSize: number) {
  return PAGE_SIZE_OPTIONS.includes(pageSize as (typeof PAGE_SIZE_OPTIONS)[number]) ? pageSize : DEFAULT_PAGE_SIZE
}

function normalizeSorting(field: string, direction: string): SandboxSorting {
  const sortField = SANDBOX_SORT_FIELDS.includes(field as SandboxListSortField)
    ? (field as SandboxListSortField)
    : DEFAULT_SANDBOX_SORTING.field
  const sortDirection = SANDBOX_SORT_DIRECTIONS.includes(direction as SandboxListSortDirection)
    ? (direction as SandboxListSortDirection)
    : DEFAULT_SANDBOX_SORTING.direction

  return {
    field: sortField,
    direction: sortDirection,
  }
}

function getValidatedStates(states: string[]): SandboxState[] {
  return states.filter((state): state is SandboxState => SANDBOX_STATES.includes(state as SandboxState))
}

function getValidatedSandboxClasses(classes: string[]): SandboxClass[] {
  return classes.filter((c): c is SandboxClass => SANDBOX_CLASSES.includes(c as SandboxClass))
}

function getNonEmptyLabels(labels: Record<string, string>) {
  return Object.fromEntries(Object.entries(labels).filter(([key, value]) => key.trim() && value.trim()))
}

function isDefaultSorting(sorting: SandboxSorting) {
  return sorting.field === DEFAULT_SANDBOX_SORTING.field && sorting.direction === DEFAULT_SANDBOX_SORTING.direction
}

function getUnknownErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message
  }

  if (typeof error === 'object' && error !== null) {
    const response = (error as { response?: { data?: { message?: unknown } } }).response
    const responseMessage = response?.data?.message
    if (typeof responseMessage === 'string') {
      return responseMessage
    }
  }

  return String(error)
}

interface UseSandboxesPageWsSyncOptions {
  currentSandboxIds: ReadonlySet<string>
  updateSandboxInCache: (sandboxId: string, updates: Partial<Sandbox>) => unknown
  markAllSandboxQueriesAsStale: (shouldRefetchActiveQueries?: boolean) => Promise<unknown>
}

function useSandboxesPageWsSync({
  currentSandboxIds,
  updateSandboxInCache,
  markAllSandboxQueriesAsStale,
}: UseSandboxesPageWsSyncOptions) {
  const revalidationTimeoutRef = useRef<number | null>(null)

  const scheduleSandboxListRevalidation = useCallback(() => {
    if (revalidationTimeoutRef.current !== null) {
      window.clearTimeout(revalidationTimeoutRef.current)
    }

    revalidationTimeoutRef.current = window.setTimeout(() => {
      revalidationTimeoutRef.current = null
      markAllSandboxQueriesAsStale(true).catch((error) => {
        console.error('Failed to revalidate sandbox list after sync', error)
      })
    }, SANDBOX_LIST_REVALIDATION_DEBOUNCE_MS)
  }, [markAllSandboxQueriesAsStale])

  useEffect(() => {
    return () => {
      if (revalidationTimeoutRef.current !== null) {
        window.clearTimeout(revalidationTimeoutRef.current)
      }
    }
  }, [])

  const reconcileSandboxListAfterSync = useCallback(
    (sandboxId: string) => {
      if (!currentSandboxIds.has(sandboxId)) {
        scheduleSandboxListRevalidation()
        return
      }

      markAllSandboxQueriesAsStale().catch((error) => {
        console.error('Failed to mark sandbox list stale after sync', error)
      })
    },
    [currentSandboxIds, markAllSandboxQueriesAsStale, scheduleSandboxListRevalidation],
  )

  useSandboxWsSync({
    onSync: (event: SandboxWsSyncEvent) => {
      if (event.type === 'created') {
        scheduleSandboxListRevalidation()
        return
      }

      if (event.type === 'state.updated') {
        let updatedState = event.newState

        if (
          event.sandbox.desiredState === SandboxDesiredState.DESTROYED &&
          (event.newState === SandboxState.ERROR || event.newState === SandboxState.BUILD_FAILED)
        ) {
          updatedState = SandboxState.DESTROYED
        }

        updateSandboxInCache(event.sandbox.id, { ...event.sandbox, state: updatedState })
        reconcileSandboxListAfterSync(event.sandbox.id)
        return
      }

      if (
        event.newDesiredState === SandboxDesiredState.DESTROYED &&
        (event.sandbox.state === SandboxState.ERROR || event.sandbox.state === SandboxState.BUILD_FAILED)
      ) {
        updateSandboxInCache(event.sandbox.id, { ...event.sandbox, state: SandboxState.DESTROYED })
      } else {
        const sandboxWithoutState: Partial<Sandbox> = { ...event.sandbox }
        delete sandboxWithoutState.state
        updateSandboxInCache(event.sandbox.id, sandboxWithoutState)
      }

      reconcileSandboxListAfterSync(event.sandbox.id)
    },
  })
}

const Sandboxes: React.FC = () => {
  const { sandboxApi, apiKeyApi, toolboxApi } = useApi()
  const { user } = useAuth()
  const navigate = useNavigate()
  const config = useConfig()
  const queryClient = useQueryClient()
  const { selectedOrganization, authenticatedUserOrganizationMember, authenticatedUserHasPermission } =
    useSelectedOrganization()

  const createSandboxSheetRef = useRef<{ open: () => undefined }>(null)
  const sandboxSheetRef = useRef<SandboxSheetRef>(null)
  const sandboxTableRef = useRef<SandboxTableRef>(null)

  const [viewParams, setViewParams] = useQueryStates(sandboxViewSearchParams)
  const [sandboxIdParam, setSandboxIdParam] = useQueryState('sandboxId', parseAsString)
  const [sandboxTabParam, setSandboxTabParam] = useQueryState('tab', tabParser)

  const pageSize = normalizePageSize(viewParams.limit)
  const [cursor, setCursor] = useState<string | undefined>(undefined)
  const [cursorHistory, setCursorHistory] = useState<(string | undefined)[]>([])

  const resetCursor = useCallback(() => {
    setCursor(undefined)
    setCursorHistory([])
  }, [])

  const filters = useMemo<SandboxFilters>(() => {
    const nextFilters: SandboxFilters = {}
    const search = viewParams.search.trim()
    const states = getValidatedStates(viewParams.states)
    const labels = getNonEmptyLabels(viewParams.labels)
    const sandboxClasses = getValidatedSandboxClasses(viewParams.sandboxClasses)

    if (search) nextFilters.name = search
    if (states.length > 0) nextFilters.states = states
    if (sandboxClasses.length > 0) nextFilters.sandboxClasses = sandboxClasses
    if (viewParams.snapshots.length > 0) nextFilters.snapshots = viewParams.snapshots
    if (viewParams.regions.length > 0) nextFilters.regions = viewParams.regions
    if (Object.keys(labels).length > 0) nextFilters.labels = labels
    if (viewParams.minCpu !== null) nextFilters.minCpu = viewParams.minCpu
    if (viewParams.maxCpu !== null) nextFilters.maxCpu = viewParams.maxCpu
    if (viewParams.minMemory !== null) nextFilters.minMemoryGib = viewParams.minMemory
    if (viewParams.maxMemory !== null) nextFilters.maxMemoryGib = viewParams.maxMemory
    if (viewParams.minDisk !== null) nextFilters.minDiskGib = viewParams.minDisk
    if (viewParams.maxDisk !== null) nextFilters.maxDiskGib = viewParams.maxDisk
    if (viewParams.lastEventAfter) nextFilters.lastEventAfter = viewParams.lastEventAfter
    if (viewParams.lastEventBefore) nextFilters.lastEventBefore = viewParams.lastEventBefore
    if (viewParams.createdAtAfter) nextFilters.createdAtAfter = viewParams.createdAtAfter
    if (viewParams.createdAtBefore) nextFilters.createdAtBefore = viewParams.createdAtBefore
    if (viewParams.isPublic !== null) nextFilters.isPublic = viewParams.isPublic
    if (viewParams.isRecoverable !== null) nextFilters.isRecoverable = viewParams.isRecoverable

    return nextFilters
  }, [
    viewParams.createdAtAfter,
    viewParams.createdAtBefore,
    viewParams.isPublic,
    viewParams.isRecoverable,
    viewParams.labels,
    viewParams.lastEventAfter,
    viewParams.lastEventBefore,
    viewParams.maxCpu,
    viewParams.maxDisk,
    viewParams.maxMemory,
    viewParams.minCpu,
    viewParams.minDisk,
    viewParams.minMemory,
    viewParams.regions,
    viewParams.sandboxClasses,
    viewParams.search,
    viewParams.snapshots,
    viewParams.states,
  ])

  const sorting = useMemo<SandboxSorting>(
    () => normalizeSorting(viewParams.sort, viewParams.order),
    [viewParams.order, viewParams.sort],
  )

  const viewResetKey = useMemo(
    () =>
      JSON.stringify({
        limit: pageSize,
        filters: {
          ...filters,
          createdAtAfter: filters.createdAtAfter?.toISOString(),
          createdAtBefore: filters.createdAtBefore?.toISOString(),
          lastEventAfter: filters.lastEventAfter?.toISOString(),
          lastEventBefore: filters.lastEventBefore?.toISOString(),
        },
        sorting,
      }),
    [filters, pageSize, sorting],
  )
  const previousViewResetKeyRef = useRef(viewResetKey)

  useEffect(() => {
    if (previousViewResetKeyRef.current === viewResetKey) {
      return
    }

    previousViewResetKeyRef.current = viewResetKey
    resetCursor()
    sandboxTableRef.current?.table.resetRowSelection()
  }, [resetCursor, viewResetKey])

  const handleNextPage = useCallback(
    (nextCursor: string | null) => {
      if (nextCursor) {
        setCursorHistory((prev) => [...prev, cursor])
        setCursor(nextCursor)
      }
    },
    [cursor],
  )

  const handlePreviousPage = useCallback(() => {
    if (cursorHistory.length > 0) {
      const newHistory = [...cursorHistory]
      const previousCursor = newHistory.pop()
      setCursorHistory(newHistory)
      setCursor(previousCursor)
    }
  }, [cursorHistory])

  const handlePageSizeChange = useCallback(
    (newPageSize: number) => {
      const nextPageSize = normalizePageSize(newPageSize)
      setViewParams({ limit: nextPageSize === DEFAULT_PAGE_SIZE ? null : nextPageSize })
      resetCursor()
    },
    [resetCursor, setViewParams],
  )

  const handleFiltersChange = useCallback(
    (newFilters: SandboxFilters) => {
      const labels = newFilters.labels ? getNonEmptyLabels(newFilters.labels) : {}

      setViewParams({
        search: newFilters.name?.trim() || null,
        states: newFilters.states?.length ? newFilters.states : null,
        snapshots: newFilters.snapshots?.length ? newFilters.snapshots : null,
        regions: newFilters.regions?.length ? newFilters.regions : null,
        sandboxClasses: newFilters.sandboxClasses?.length ? newFilters.sandboxClasses : null,
        labels: Object.keys(labels).length > 0 ? labels : null,
        minCpu: newFilters.minCpu ?? null,
        maxCpu: newFilters.maxCpu ?? null,
        minMemory: newFilters.minMemoryGib ?? null,
        maxMemory: newFilters.maxMemoryGib ?? null,
        minDisk: newFilters.minDiskGib ?? null,
        maxDisk: newFilters.maxDiskGib ?? null,
        lastEventAfter: newFilters.lastEventAfter ?? null,
        lastEventBefore: newFilters.lastEventBefore ?? null,
        createdAtAfter: newFilters.createdAtAfter ?? null,
        createdAtBefore: newFilters.createdAtBefore ?? null,
        isPublic: newFilters.isPublic ?? null,
        isRecoverable: newFilters.isRecoverable ?? null,
      })
      resetCursor()
    },
    [resetCursor, setViewParams],
  )

  const handleSortingChange = useCallback(
    (newSorting: SandboxSorting) => {
      const nextSorting = {
        field: newSorting.field ?? DEFAULT_SANDBOX_SORTING.field,
        direction: newSorting.direction ?? DEFAULT_SANDBOX_SORTING.direction,
      }

      setViewParams({
        sort: isDefaultSorting(nextSorting) ? null : (nextSorting.field ?? null),
        order: isDefaultSorting(nextSorting) ? null : (nextSorting.direction ?? null),
      })
      resetCursor()
    },
    [resetCursor, setViewParams],
  )

  const queryParams = useMemo<SandboxQueryParams>(
    () => ({
      cursor,
      limit: pageSize,
      filters,
      sorting,
    }),
    [cursor, filters, pageSize, sorting],
  )

  const sandboxListQueryKey = useMemo(
    () => queryKeys.sandboxes.list(selectedOrganization?.id ?? ''),
    [selectedOrganization?.id],
  )
  const queryKey = useMemo(
    () => queryKeys.sandboxes.list(selectedOrganization?.id ?? '', queryParams),
    [queryParams, selectedOrganization?.id],
  )

  const {
    data: sandboxesData,
    isLoading: sandboxesDataIsLoading,
    isFetching: sandboxesDataIsFetching,
    isPlaceholderData: sandboxesDataIsPlaceholderData,
    error: sandboxesDataError,
    refetch: refetchSandboxesData,
  } = useSandboxesQuery(queryParams)

  const sandboxes = sandboxesData?.items || DEFAULT_SANDBOXES
  const currentSandboxIds = useMemo(() => new Set(sandboxes.map((sandbox) => sandbox.id)), [sandboxes])

  useEffect(() => {
    if (sandboxesDataError) {
      handleApiError(sandboxesDataError, 'Failed to fetch sandboxes')
    }
  }, [sandboxesDataError])

  const [sandboxToDelete, setSandboxToDelete] = useState<string | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [forkTreeSandboxId, setForkTreeSandboxId] = useState<string | null>(null)
  const [recursiveDeleteSandboxId, setRecursiveDeleteSandboxId] = useState<string | null>(null)
  const [sandboxToSnapshot, setSandboxToSnapshot] = useState<string | null>(null)
  const [snapshotName, setSnapshotName] = useState('')
  const [snapshotIncludeMemory, setSnapshotIncludeMemory] = useState(false)
  const [selectedSandbox, setSelectedSandbox] = useState<SandboxListItem | null>(null)
  const [orderedSandboxItems, setOrderedSandboxItems] = useState<SandboxListItem[] | null>(null)
  const [showCreateSshDialog, setShowCreateSshDialog] = useState(false)
  const [showRevokeSshDialog, setShowRevokeSshDialog] = useState(false)
  const [sshSandboxId, setSshSandboxId] = useState('')

  const seedSandboxDetailsCache = useCallback(
    (sandbox: SandboxListItem | Sandbox) => {
      if (!selectedOrganization?.id) {
        return
      }

      const queryKey = queryKeys.sandboxes.detail(selectedOrganization.id, sandbox.id)
      if (queryClient.getQueryData<Sandbox>(queryKey)) {
        return
      }

      queryClient.setQueryData<Sandbox>(queryKey, sandbox as Sandbox, { updatedAt: 0 })
    },
    [queryClient, selectedOrganization?.id],
  )

  const updateSandboxInCache = useCallback(
    (sandboxId: string, updates: Partial<Sandbox>) => {
      queryClient.setQueryData<Sandbox>(
        queryKeys.sandboxes.detail(selectedOrganization?.id ?? '', sandboxId),
        (oldData) => (oldData ? { ...oldData, ...updates } : oldData),
      )

      queryClient.setQueriesData<ListSandboxesResponse>({ queryKey: sandboxListQueryKey }, (oldData) => {
        if (!oldData) return oldData

        return {
          ...oldData,
          items: (oldData.items || DEFAULT_SANDBOXES).map((sandbox) =>
            sandbox.id === sandboxId ? { ...sandbox, ...updates } : sandbox,
          ),
        }
      })

      setSelectedSandbox((prev) => (prev?.id === sandboxId ? { ...prev, ...updates } : prev))
    },
    [queryClient, sandboxListQueryKey, selectedOrganization?.id],
  )

  const markAllSandboxQueriesAsStale = useCallback(
    async (shouldRefetchActiveQueries = false) => {
      await queryClient.invalidateQueries({
        queryKey: sandboxListQueryKey,
        refetchType: shouldRefetchActiveQueries ? 'active' : 'none',
      })
    },
    [queryClient, sandboxListQueryKey],
  )

  const cancelCurrentSandboxQueryRefetches = useCallback(async () => {
    await queryClient.cancelQueries({ queryKey })
  }, [queryClient, queryKey])

  useSandboxesPageWsSync({
    currentSandboxIds,
    updateSandboxInCache,
    markAllSandboxQueriesAsStale,
  })

  useEffect(() => {
    if (!sandboxesDataIsPlaceholderData && sandboxes.length === 0 && cursorHistory.length > 0) {
      handlePreviousPage()
    }
  }, [cursorHistory.length, handlePreviousPage, sandboxes.length, sandboxesDataIsPlaceholderData])

  const pendingSandboxIds = useMutatingSandboxes()
  const sandboxIsLoading = useMemo(() => {
    return Object.fromEntries(Array.from(pendingSandboxIds).map((sandboxId) => [sandboxId, true]))
  }, [pendingSandboxIds])

  const startSandboxMutation = useStartSandboxMutation({ invalidate: false })
  const stopSandboxMutation = useStopSandboxMutation({ invalidate: false })
  const pauseSandboxMutation = usePauseSandboxMutation({ invalidate: false })
  const archiveSandboxMutation = useArchiveSandboxMutation({ invalidate: false })
  const recoverSandboxMutation = useRecoverSandboxMutation({ invalidate: false })
  const deleteSandboxMutation = useDeleteSandboxMutation({ invalidate: false })

  const forkSandboxMutation = useMutation({
    mutationKey: mutationKeys.sandboxes.fork(),
    mutationFn: async ({ sandboxId }: { sandboxId: string }) => {
      await sandboxApi.forkSandbox(sandboxId, {}, selectedOrganization?.id)
    },
  })

  const createSandboxSnapshotMutation = useMutation({
    mutationKey: mutationKeys.sandboxes.createSnapshot(),
    mutationFn: async ({ sandboxId, name, includeMemory }: CreateSandboxSnapshotVariables) => {
      await sandboxApi.createSandboxSnapshot(sandboxId, { name, includeMemory }, selectedOrganization?.id)
    },
  })

  const getSandboxById = useCallback(
    (sandboxId: string) => {
      const cachedSandbox = selectedOrganization?.id
        ? queryClient.getQueryData<Sandbox>(queryKeys.sandboxes.detail(selectedOrganization.id, sandboxId))
        : undefined

      return (
        sandboxes.find((sandbox) => sandbox.id === sandboxId) ??
        cachedSandbox ??
        (selectedSandbox?.id === sandboxId ? selectedSandbox : undefined)
      )
    },
    [queryClient, sandboxes, selectedOrganization?.id, selectedSandbox],
  )

  const getPortPreviewUrl = useCallback(
    async (sandboxId: string, port: number): Promise<string> => {
      return (await sandboxApi.getSignedPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url
    },
    [sandboxApi, selectedOrganization?.id],
  )

  const getVncUrl = useCallback(
    async (sandboxId: string): Promise<string | null> => {
      try {
        const portPreviewUrl = await getPortPreviewUrl(sandboxId, 6080)
        return `${portPreviewUrl}/vnc.html`
      } catch (error) {
        handleApiError(error, 'Failed to construct VNC URL')
        return null
      }
    },
    [getPortPreviewUrl],
  )

  const vncMutation = useMutation({
    mutationKey: mutationKeys.sandboxes.vnc(),
    mutationFn: async ({ sandboxId }: { sandboxId: string }) => {
      toast.info('Checking VNC desktop status...')

      try {
        const statusResponse = await toolboxApi.getComputerUseStatusDeprecated(sandboxId, selectedOrganization?.id)
        const status = statusResponse.data.status

        if (status === 'active') {
          const vncUrl = await getVncUrl(sandboxId)
          if (vncUrl) {
            window.open(vncUrl, '_blank')
            toast.success('Opening VNC desktop...')
          }
          return
        }

        try {
          await toolboxApi.startComputerUseDeprecated(sandboxId, selectedOrganization?.id)
          toast.success('Starting VNC desktop...')
          await new Promise((resolve) => setTimeout(resolve, 5000))

          const newStatusResponse = await toolboxApi.getComputerUseStatusDeprecated(sandboxId, selectedOrganization?.id)
          const newStatus = newStatusResponse.data.status

          if (newStatus === 'active') {
            const vncUrl = await getVncUrl(sandboxId)
            if (vncUrl) {
              window.open(vncUrl, '_blank')
              toast.success('VNC desktop is ready!', {
                action: (
                  <Button variant="secondary" onClick={() => window.open(vncUrl, '_blank')}>
                    Open in new tab
                  </Button>
                ),
              })
            }
          } else {
            toast.error(`VNC desktop failed to start. Status: ${newStatus}`)
          }
        } catch (startError) {
          const errorMessage = getUnknownErrorMessage(startError)

          if (errorMessage === 'Computer-use functionality is not available') {
            toast.error('Computer-use functionality is not available', {
              description: (
                <div>
                  <div>Computer-use dependencies are missing in the runtime environment.</div>
                  <div className="mt-2">
                    <a
                      href={`${DAYTONA_DOCS_URL}/getting-started/computer-use`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                    >
                      See documentation on how to configure the runtime for computer-use
                    </a>
                  </div>
                </div>
              ),
            })
          } else {
            handleApiError(startError, 'Failed to start VNC desktop')
          }
        }
      } catch (error) {
        handleApiError(error, 'Failed to check VNC status')
      }
    },
  })

  const screenRecordingsMutation = useMutation({
    mutationKey: mutationKeys.sandboxes.screenRecordings(),
    mutationFn: async ({ sandboxId }: { sandboxId: string }) => {
      const sandbox = getSandboxById(sandboxId)
      if (!sandbox || sandbox.state !== SandboxState.STARTED) {
        toast.error('Sandbox must be started to access Screen Recordings')
        return
      }

      const portPreviewUrl = await getPortPreviewUrl(sandboxId, 33333)
      window.open(portPreviewUrl, '_blank')
      toast.success('Opening Screen Recordings dashboard...')
    },
  })

  const performSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, newState: SandboxState) => {
      updateSandboxInCache(sandboxId, { state: newState })
    },
    [updateSandboxInCache],
  )

  const revertSandboxStateOptimisticUpdate = useCallback(
    (sandboxId: string, previousState?: SandboxState) => {
      if (!previousState) {
        return
      }

      updateSandboxInCache(sandboxId, { state: previousState })
    },
    [updateSandboxInCache],
  )

  const handleRefresh = useCallback(async () => {
    try {
      await refetchSandboxesData()
    } catch (error) {
      handleApiError(error, 'Failed to refresh sandboxes')
    }
  }, [refetchSandboxesData])

  const [snapshotFilters, setSnapshotFilters] = useState<SnapshotFilters>({})

  const handleSnapshotFiltersChange = useCallback((snapshotFilterUpdate: Partial<SnapshotFilters>) => {
    setSnapshotFilters((prev) => ({ ...prev, ...snapshotFilterUpdate }))
  }, [])

  const snapshotsQueryParams = useMemo<SnapshotQueryParams>(
    () => ({
      page: 1,
      pageSize: 100,
      filters: snapshotFilters,
    }),
    [snapshotFilters],
  )

  const {
    data: snapshotsData,
    isLoading: snapshotsDataIsLoading,
    error: snapshotsDataError,
  } = useSnapshotsQuery(snapshotsQueryParams)

  const snapshotsDataHasMore = useMemo(() => {
    return snapshotsData && snapshotsData.totalPages > 1
  }, [snapshotsData])

  useEffect(() => {
    if (snapshotsDataError) {
      handleApiError(snapshotsDataError, 'Failed to fetch snapshots')
    }
  }, [snapshotsDataError])

  const { data: regionsData = EMPTY_REGIONS, isLoading: regionsDataIsLoading } = useAvailableRegionsQuery(
    selectedOrganization?.id,
  )
  const { getRegionName } = useRegionLookup(selectedOrganization?.id)

  const sandboxFromLoadedResults = useMemo(
    () => sandboxes.find((sandbox) => sandbox.id === sandboxIdParam),
    [sandboxIdParam, sandboxes],
  )

  useEffect(() => {
    if (!sandboxIdParam) {
      setSelectedSandbox(null)
      setOrderedSandboxItems(null)
      sandboxSheetRef.current?.close()
      return
    }

    if (sandboxFromLoadedResults) {
      seedSandboxDetailsCache(sandboxFromLoadedResults)
    }
    setSelectedSandbox(sandboxFromLoadedResults ?? null)
    sandboxSheetRef.current?.open()
  }, [sandboxFromLoadedResults, sandboxIdParam, seedSandboxDetailsCache])

  const handleCreateSnapshot = (id: string) => {
    const sandbox = sandboxes.find((s) => s.id === id)
    setSandboxToSnapshot(id)
    setSnapshotName('')
    setSnapshotIncludeMemory(sandbox?.sandboxClass === SandboxClass.WINDOWS && sandbox?.state === SandboxState.STARTED)
  }

  const handleFork = async (id: string) => {
    try {
      await forkSandboxMutation.mutateAsync({ sandboxId: id })
      toast.success('Fork started')
      await markAllSandboxQueriesAsStale(true)
    } catch (error) {
      handleApiError(error, 'Failed to fork sandbox')
    }
  }

  const handleViewForks = (id: string) => {
    setForkTreeSandboxId(id)
  }

  const openDeleteDialog = async (id: string) => {
    try {
      const forksRes = await sandboxApi.getSandboxForks(id, selectedOrganization?.id)
      if (forksRes.data.length > 0) {
        setRecursiveDeleteSandboxId(id)
        return
      }
    } catch {
      // Fall through to normal delete if fork check fails.
    }
    setSandboxToDelete(id)
    setShowDeleteDialog(true)
  }

  const handleStart = async (id: string) => {
    const sandboxToStart = getSandboxById(id)
    const previousState = sandboxToStart?.state
    const wasPaused = previousState === SandboxState.PAUSED

    await cancelCurrentSandboxQueryRefetches()
    const optimisticStartState =
      previousState === SandboxState.ARCHIVED
        ? SandboxState.RESTORING
        : wasPaused
          ? SandboxState.RESUMING
          : SandboxState.STARTING
    performSandboxStateOptimisticUpdate(id, optimisticStartState)

    try {
      await startSandboxMutation.mutateAsync({ sandboxId: id })
      toast.success(`${wasPaused ? 'Resuming' : 'Starting'} sandbox with ID: ${id}`)
      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, `Failed to ${wasPaused ? 'resume' : 'start'} sandbox`, {
        action:
          error instanceof OrganizationSuspendedError &&
          config.billingApiUrl &&
          authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER ? (
            <Button variant="secondary" onClick={() => navigate(RoutePath.BILLING_WALLET)}>
              Go to billing
            </Button>
          ) : null,
      })
      revertSandboxStateOptimisticUpdate(id, previousState)
    }
  }

  const handleRecover = async (id: string) => {
    const sandboxToRecover = getSandboxById(id)
    const previousState = sandboxToRecover?.state

    await cancelCurrentSandboxQueryRefetches()
    performSandboxStateOptimisticUpdate(id, SandboxState.STARTING)

    try {
      await recoverSandboxMutation.mutateAsync({ sandboxId: id })
      toast.success('Sandbox recovered. Restarting...')
      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to recover sandbox')
      revertSandboxStateOptimisticUpdate(id, previousState)
    }
  }

  const handleStop = async (id: string) => {
    const sandboxToStop = getSandboxById(id)
    const previousState = sandboxToStop?.state

    await cancelCurrentSandboxQueryRefetches()
    performSandboxStateOptimisticUpdate(id, SandboxState.STOPPING)

    try {
      await stopSandboxMutation.mutateAsync({ sandboxId: id })
      toast.success(
        `Stopping sandbox with ID: ${id}`,
        sandboxToStop?.autoDeleteInterval !== undefined && sandboxToStop.autoDeleteInterval >= 0
          ? {
              description: `This sandbox will be deleted automatically ${sandboxToStop.autoDeleteInterval === 0 ? 'upon stopping' : `in ${formatDuration(sandboxToStop.autoDeleteInterval)} unless it is started again`}.`,
            }
          : undefined,
      )
      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to stop sandbox')
      revertSandboxStateOptimisticUpdate(id, previousState)
    }
  }

  const handlePause = async (id: string) => {
    const sandboxToPause = getSandboxById(id)
    const previousState = sandboxToPause?.state

    await cancelCurrentSandboxQueryRefetches()
    performSandboxStateOptimisticUpdate(id, SandboxState.PAUSING)

    try {
      await pauseSandboxMutation.mutateAsync({ sandboxId: id })
      toast.success(`Pausing sandbox with ID: ${id}`)
      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to pause sandbox')
      revertSandboxStateOptimisticUpdate(id, previousState)
    }
  }

  const handleDelete = async (id: string) => {
    const sandboxToDeleteItem = getSandboxById(id)
    const previousState = sandboxToDeleteItem?.state

    await cancelCurrentSandboxQueryRefetches()
    performSandboxStateOptimisticUpdate(id, SandboxState.DESTROYING)

    try {
      await deleteSandboxMutation.mutateAsync({ sandboxId: id })
      setSandboxToDelete(null)
      setShowDeleteDialog(false)

      if (sandboxIdParam === id) {
        setSandboxIdParam(null)
        setSandboxTabParam(null)
      }

      toast.success(`Deleting sandbox with ID: ${id}`)
      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to delete sandbox')
      revertSandboxStateOptimisticUpdate(id, previousState)
    }
  }

  const handleArchive = async (id: string) => {
    const sandboxToArchive = getSandboxById(id)
    const previousState = sandboxToArchive?.state

    await cancelCurrentSandboxQueryRefetches()
    performSandboxStateOptimisticUpdate(id, SandboxState.ARCHIVING)

    try {
      await archiveSandboxMutation.mutateAsync({ sandboxId: id })
      toast.success(`Archiving sandbox with ID: ${id}`)
      await markAllSandboxQueriesAsStale()
    } catch (error) {
      handleApiError(error, 'Failed to archive sandbox')
      revertSandboxStateOptimisticUpdate(id, previousState)
    }
  }

  const executeBulkAction = useCallback(
    async ({
      ids,
      actionName,
      optimisticState,
      apiCall,
      toastMessages,
    }: {
      ids: string[]
      actionName: string
      optimisticState: SandboxState | ((previousState: SandboxState | undefined) => SandboxState)
      apiCall: (id: string) => Promise<unknown>
      toastMessages: {
        successTitle: string
        errorTitle: string
        warningTitle: string
        canceledTitle: string
      }
    }) => {
      await cancelCurrentSandboxQueryRefetches()

      const previousStatesById = new Map(sandboxes.map((sandbox) => [sandbox.id, sandbox.state]))

      let isCancelled = false
      let processedCount = 0
      let successCount = 0
      let failureCount = 0

      const totalLabel = pluralize(ids.length, 'sandbox', 'sandboxes')
      const onCancel = () => {
        isCancelled = true
      }

      const bulkToast = createBulkActionToast(`${actionName} 0 of ${totalLabel}.`, {
        action: { label: 'Cancel', onClick: onCancel },
      })

      try {
        for (const id of ids) {
          if (isCancelled) break

          processedCount += 1
          bulkToast.loading(`${actionName} ${processedCount} of ${totalLabel}.`, {
            action: { label: 'Cancel', onClick: onCancel },
          })

          const resolvedOptimisticState =
            typeof optimisticState === 'function' ? optimisticState(previousStatesById.get(id)) : optimisticState
          performSandboxStateOptimisticUpdate(id, resolvedOptimisticState)

          try {
            await apiCall(id)
            successCount += 1
          } catch (error) {
            failureCount += 1
            revertSandboxStateOptimisticUpdate(id, previousStatesById.get(id))
            console.error(`${actionName} sandbox failed`, id, error)
          }
        }

        await markAllSandboxQueriesAsStale()
        bulkToast.result({ successCount, failureCount }, toastMessages)
      } catch (error) {
        console.error(`${actionName} sandboxes failed`, error)
        bulkToast.error(`${actionName} sandboxes failed.`)
      }

      return { successCount, failureCount }
    },
    [
      cancelCurrentSandboxQueryRefetches,
      markAllSandboxQueriesAsStale,
      performSandboxStateOptimisticUpdate,
      revertSandboxStateOptimisticUpdate,
      sandboxes,
    ],
  )

  const handleBulkStart = (ids: string[]) =>
    executeBulkAction({
      ids,
      actionName: 'Starting',
      optimisticState: (previousState) =>
        previousState === SandboxState.ARCHIVED ? SandboxState.RESTORING : SandboxState.STARTING,
      apiCall: (id) => startSandboxMutation.mutateAsync({ sandboxId: id }),
      toastMessages: {
        successTitle: `${pluralize(ids.length, 'sandbox', 'sandboxes')} started.`,
        errorTitle: `Failed to start ${pluralize(ids.length, 'sandbox', 'sandboxes')}.`,
        warningTitle: 'Failed to start some sandboxes.',
        canceledTitle: 'Start canceled.',
      },
    })

  const handleBulkStop = (ids: string[]) =>
    executeBulkAction({
      ids,
      actionName: 'Stopping',
      optimisticState: SandboxState.STOPPING,
      apiCall: (id) => stopSandboxMutation.mutateAsync({ sandboxId: id }),
      toastMessages: {
        successTitle: `${pluralize(ids.length, 'sandbox', 'sandboxes')} stopped.`,
        errorTitle: `Failed to stop ${pluralize(ids.length, 'sandbox', 'sandboxes')}.`,
        warningTitle: 'Failed to stop some sandboxes.',
        canceledTitle: 'Stop canceled.',
      },
    })

  const handleBulkArchive = (ids: string[]) =>
    executeBulkAction({
      ids,
      actionName: 'Archiving',
      optimisticState: SandboxState.ARCHIVING,
      apiCall: (id) => archiveSandboxMutation.mutateAsync({ sandboxId: id }),
      toastMessages: {
        successTitle: `${pluralize(ids.length, 'sandbox', 'sandboxes')} archived.`,
        errorTitle: `Failed to archive ${pluralize(ids.length, 'sandbox', 'sandboxes')}.`,
        warningTitle: 'Failed to archive some sandboxes.',
        canceledTitle: 'Archive canceled.',
      },
    })

  const handleBulkDelete = async (ids: string[]) => {
    const selectedSandboxInBulk = sandboxIdParam ? ids.includes(sandboxIdParam) : false

    await executeBulkAction({
      ids,
      actionName: 'Deleting',
      optimisticState: SandboxState.DESTROYING,
      apiCall: (id) => deleteSandboxMutation.mutateAsync({ sandboxId: id }),
      toastMessages: {
        successTitle: `${pluralize(ids.length, 'sandbox', 'sandboxes')} deleted.`,
        errorTitle: `Failed to delete ${pluralize(ids.length, 'sandbox', 'sandboxes')}.`,
        warningTitle: 'Failed to delete some sandboxes.',
        canceledTitle: 'Delete canceled.',
      },
    })

    if (selectedSandboxInBulk) {
      setSandboxIdParam(null)
      setSandboxTabParam(null)
    }
  }

  const handleVnc = async (id: string) => {
    await vncMutation.mutateAsync({ sandboxId: id })
  }

  const handleScreenRecordings = async (id: string) => {
    try {
      await screenRecordingsMutation.mutateAsync({ sandboxId: id })
    } catch (error) {
      handleApiError(error, 'Failed to open Screen Recordings')
    }
  }

  const openCreateSshDialog = (id: string) => {
    setSshSandboxId(id)
    setShowCreateSshDialog(true)
  }

  const openRevokeSshDialog = (id: string) => {
    setSshSandboxId(id)
    setShowRevokeSshDialog(true)
  }

  const sandboxItems = useMemo(() => orderedSandboxItems ?? sandboxes, [orderedSandboxItems, sandboxes])
  const selectedSandboxIndex = useMemo(
    () => sandboxItems.findIndex((sandbox) => sandbox.id === sandboxIdParam),
    [sandboxIdParam, sandboxItems],
  )

  const handleSandboxSheetNavigate = (direction: 'prev' | 'next') => {
    if (selectedSandboxIndex < 0) {
      return
    }

    const nextIndex = direction === 'prev' ? selectedSandboxIndex - 1 : selectedSandboxIndex + 1
    const nextSandbox = sandboxItems[nextIndex]

    if (nextSandbox) {
      seedSandboxDetailsCache(nextSandbox)
      setSelectedSandbox(nextSandbox)
      setSandboxIdParam(nextSandbox.id)
    }
  }

  const handleSandboxDetailsOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      setSandboxIdParam(null)
      setSandboxTabParam(null)
    }
  }

  const openSandboxDetails = (sandbox: SandboxListItem, defaultTab: SandboxDetailsSheetTabValue = 'overview') => {
    const orderedSandboxes =
      sandboxTableRef.current?.table.getPrePaginationRowModel().rows.map((row) => row.original) ?? []
    seedSandboxDetailsCache(sandbox)
    setOrderedSandboxItems(orderedSandboxes.some((item) => item.id === sandbox.id) ? orderedSandboxes : null)
    setSelectedSandbox(sandbox)
    setSandboxTabParam(defaultTab)
    setSandboxIdParam(sandbox.id)
  }

  const handleSandboxRowClick = (sandbox: SandboxListItem) => {
    openSandboxDetails(sandbox)
  }

  const handleOpenTerminal = (sandbox: SandboxListItem) => {
    openSandboxDetails(sandbox, 'terminal')
  }

  const handleSandboxCreated = (sandbox: CreatedSandbox) => {
    const createdSandbox = sandbox as unknown as Sandbox

    markAllSandboxQueriesAsStale(true)
    openSandboxDetails(createdSandbox)
  }

  const writePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES),
    [authenticatedUserHasPermission],
  )
  const deletePermitted = useMemo(
    () => authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES),
    [authenticatedUserHasPermission],
  )
  const canCreateSandbox = writePermitted && !selectedOrganization?.suspended

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (!canCreateSandbox) {
      return []
    }

    return [
      {
        id: 'create-sandbox',
        label: 'Create Sandbox',
        icon: <PlusIcon className="h-4 w-4" />,
        onSelect: () => createSandboxSheetRef.current?.open(),
      },
    ]
  }, [canCreateSandbox])

  useRegisterCommands(rootCommands, { groupId: 'sandbox-actions', groupLabel: 'Sandbox actions', groupOrder: 0 })

  useEffect(() => {
    const onboardIfNeeded = async () => {
      if (!selectedOrganization) {
        return
      }

      const skipOnboardingKey = `${LocalStorageKey.SkipOnboardingPrefix}${user?.profile.sub}`
      const shouldSkipOnboarding = getLocalStorageItem(skipOnboardingKey) === 'true'

      if (shouldSkipOnboarding) {
        return
      }

      try {
        const keys = (await apiKeyApi.listApiKeys(selectedOrganization.id)).data
        if (keys.length === 0) {
          setLocalStorageItem(skipOnboardingKey, 'true')
          navigate(RoutePath.ONBOARDING)
        } else {
          setLocalStorageItem(skipOnboardingKey, 'true')
        }
      } catch (error) {
        console.error('Failed to check if user needs onboarding', error)
      }
    }

    onboardIfNeeded()
  }, [apiKeyApi, navigate, selectedOrganization, user])

  const handleCreateSnapshotConfirm = async (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault()
    if (!sandboxToSnapshot || !snapshotName.trim()) return

    try {
      await createSandboxSnapshotMutation.mutateAsync({
        sandboxId: sandboxToSnapshot,
        name: snapshotName.trim(),
        includeMemory: snapshotIncludeMemory,
      })
      toast.success('Snapshot creation started')
      setSandboxToSnapshot(null)
      setSnapshotName('')
      setSnapshotIncludeMemory(false)
    } catch (error) {
      handleApiError(error, 'Failed to create snapshot')
    }
  }

  return (
    <PageLayout contained>
      <PageHeader />
      <PageContent size="full" className="overflow-hidden">
        <PageIntro
          title="Sandboxes"
          actions={
            <>
              {!sandboxesDataIsLoading && sandboxes.length === 0 && (
                <Button
                  variant="link"
                  onClick={() => navigate(RoutePath.ONBOARDING)}
                  size="sm"
                  className="text-muted-foreground"
                >
                  Onboarding guide
                </Button>
              )}

              {canCreateSandbox && (
                <CreateSandboxSheet ref={createSandboxSheetRef} onSandboxCreated={handleSandboxCreated} />
              )}
            </>
          }
        />
        <SandboxTable
          ref={sandboxTableRef}
          sandboxIsLoading={sandboxIsLoading}
          activeSandboxId={sandboxIdParam ?? undefined}
          handleStart={handleStart}
          handleStop={handleStop}
          handleDelete={openDeleteDialog}
          handleBulkDelete={handleBulkDelete}
          handleBulkStart={handleBulkStart}
          handleBulkStop={handleBulkStop}
          handleBulkArchive={handleBulkArchive}
          handleArchive={handleArchive}
          handleVnc={handleVnc}
          handleCreateSshAccess={openCreateSshDialog}
          handleRevokeSshAccess={openRevokeSshDialog}
          handleRefresh={handleRefresh}
          isRefreshing={sandboxesDataIsFetching}
          data={sandboxes}
          loading={sandboxesDataIsLoading}
          isShowingPreviousData={sandboxesDataIsPlaceholderData}
          snapshots={snapshotsData?.items ?? []}
          snapshotsDataIsLoading={snapshotsDataIsLoading}
          snapshotsDataHasMore={snapshotsDataHasMore}
          onChangeSnapshotSearchValue={(name?: string) => handleSnapshotFiltersChange({ name })}
          regionsData={regionsData ?? []}
          regionsDataIsLoading={regionsDataIsLoading}
          onRowClick={handleSandboxRowClick}
          sorting={sorting}
          onSortingChange={handleSortingChange}
          filters={filters}
          onFiltersChange={handleFiltersChange}
          handleRecover={handleRecover}
          getRegionName={getRegionName}
          handleScreenRecordings={handleScreenRecordings}
          handleCreateSnapshot={handleCreateSnapshot}
          handleFork={handleFork}
          handleViewForks={handleViewForks}
          handlePause={handlePause}
          handleOpenTerminal={handleOpenTerminal}
        />

        {sandboxToDelete && (
          <AlertDialog
            open={showDeleteDialog}
            onOpenChange={(isOpen) => {
              if (!isOpen && sandboxIsLoading[sandboxToDelete]) {
                return
              }

              setShowDeleteDialog(isOpen)
              if (!isOpen) {
                setSandboxToDelete(null)
              }
            }}
          >
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Confirm Sandbox Deletion</AlertDialogTitle>
                <AlertDialogDescription>
                  Are you sure you want to delete this sandbox? This action cannot be undone.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel disabled={sandboxIsLoading[sandboxToDelete]}>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  variant="destructive"
                  disabled={sandboxIsLoading[sandboxToDelete]}
                  onClick={async (event) => {
                    event.preventDefault()
                    await handleDelete(sandboxToDelete)
                  }}
                >
                  {sandboxIsLoading[sandboxToDelete] ? 'Deleting...' : 'Delete'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}

        {sandboxToSnapshot && (
          <AlertDialog
            open={Boolean(sandboxToSnapshot)}
            onOpenChange={(isOpen) => {
              if (!isOpen) {
                setSandboxToSnapshot(null)
                setSnapshotName('')
                setSnapshotIncludeMemory(false)
              }
            }}
          >
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Create Snapshot</AlertDialogTitle>
                <AlertDialogDescription>Enter a name for the new snapshot.</AlertDialogDescription>
              </AlertDialogHeader>
              <Input
                value={snapshotName}
                onChange={(event) => setSnapshotName(event.target.value)}
                placeholder="Snapshot name"
                disabled={createSandboxSnapshotMutation.isPending}
              />
              {sandboxes.find((s) => s.id === sandboxToSnapshot)?.sandboxClass === SandboxClass.WINDOWS && (
                <div className="flex items-start gap-3">
                  <Checkbox id="snapshot-include-memory" checked={snapshotIncludeMemory} disabled className="mt-0.5" />
                  <div className="grid gap-1 leading-none">
                    <Label htmlFor="snapshot-include-memory" className="text-sm">
                      Include memory state
                    </Label>
                    <p className="text-muted-foreground text-xs">
                      {snapshotIncludeMemory
                        ? 'Sandbox is running — memory will be captured. Stop the sandbox first for a filesystem-only snapshot.'
                        : 'Sandbox is stopped — filesystem-only snapshot. Start the sandbox first to capture memory.'}
                    </p>
                  </div>
                </div>
              )}
              <AlertDialogFooter>
                <AlertDialogCancel disabled={createSandboxSnapshotMutation.isPending}>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  disabled={!snapshotName.trim() || createSandboxSnapshotMutation.isPending}
                  onClick={handleCreateSnapshotConfirm}
                >
                  {createSandboxSnapshotMutation.isPending ? 'Creating...' : 'Create'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}

        <CreateSshAccessSheet
          sandboxId={sshSandboxId}
          open={showCreateSshDialog}
          onOpenChange={(isOpen) => {
            setShowCreateSshDialog(isOpen)
            if (!isOpen) {
              setSshSandboxId('')
            }
          }}
        />

        <RevokeSshAccessDialog
          sandboxId={sshSandboxId}
          open={showRevokeSshDialog}
          onOpenChange={(isOpen) => {
            setShowRevokeSshDialog(isOpen)
            if (!isOpen) {
              setSshSandboxId('')
            }
          }}
        />

        <SandboxDetailsSheet
          ref={sandboxSheetRef}
          sandboxId={sandboxIdParam}
          onOpenChange={handleSandboxDetailsOpenChange}
          sandboxIsLoading={sandboxIsLoading}
          handleStart={handleStart}
          handleStop={handleStop}
          handlePause={handlePause}
          handleDelete={async (id) => {
            await openDeleteDialog(id)
          }}
          handleArchive={handleArchive}
          writePermitted={writePermitted}
          deletePermitted={deletePermitted}
          handleRecover={handleRecover}
          getRegionName={getRegionName}
          onCreateSshAccess={openCreateSshDialog}
          onRevokeSshAccess={openRevokeSshDialog}
          onScreenRecordings={handleScreenRecordings}
          onNavigate={handleSandboxSheetNavigate}
          hasPrev={selectedSandboxIndex > 0}
          hasNext={selectedSandboxIndex >= 0 && selectedSandboxIndex < sandboxItems.length - 1}
          defaultTab={sandboxTabParam ?? 'overview'}
        />

        {forkTreeSandboxId && (
          <ForkTreeDialog
            sandboxId={forkTreeSandboxId}
            open={Boolean(forkTreeSandboxId)}
            onClose={() => setForkTreeSandboxId(null)}
          />
        )}

        {recursiveDeleteSandboxId && (
          <RecursiveDeleteDialog
            sandboxId={recursiveDeleteSandboxId}
            open={Boolean(recursiveDeleteSandboxId)}
            onClose={() => setRecursiveDeleteSandboxId(null)}
            onDeleted={async () => {
              await markAllSandboxQueriesAsStale(true)
            }}
          />
        )}
      </PageContent>
      <PageFooter>
        <CursorPagination
          className="justify-between w-full"
          pageSize={pageSize}
          onPageSizeChange={handlePageSizeChange}
          hasNextPage={Boolean(sandboxesData?.nextCursor)}
          hasPreviousPage={cursorHistory.length > 0}
          onNextPage={() => handleNextPage(sandboxesData?.nextCursor ?? null)}
          onPreviousPage={handlePreviousPage}
        />
      </PageFooter>
    </PageLayout>
  )
}

export default Sandboxes
