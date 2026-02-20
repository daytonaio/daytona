/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { queryKeys } from '@/hooks/queries/queryKeys'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import {
  CreateSandboxBaseParams,
  CreateSandboxFromImageParams,
  CreateSandboxFromSnapshotParams,
  Daytona,
  Sandbox,
} from '@daytonaio/sdk'
import { QueryClient, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

type CreateSandboxParams = CreateSandboxBaseParams | CreateSandboxFromImageParams | CreateSandboxFromSnapshotParams

const TERMINAL_PORT = 22222
const VNC_PORT = 6080

export type UseSandboxSessionOptions = {
  scope?: string
  createParams?: CreateSandboxParams
  terminal?: boolean
  vnc?: boolean
  notify?: { sandbox?: boolean; terminal?: boolean; vnc?: boolean }
}

export type SandboxState = {
  instance: Sandbox | null
  loading: boolean
  error: string | null
  create: (params?: CreateSandboxParams) => Promise<Sandbox>
}

export type PortQueryState = {
  url: string | null
  loading: boolean
  error: string | null
  refetch: () => void
}

export type VncState = PortQueryState & {
  start: () => void
}

export type UseSandboxSessionResult = {
  sandbox: SandboxState
  terminal: PortQueryState
  vnc: VncState
}

export function removeSandboxSessionQueries(queryClient: QueryClient, scope: string): void {
  queryClient
    .getMutationCache()
    .findAll({ mutationKey: ['create-sandbox', scope] })
    .forEach((m) => queryClient.getMutationCache().remove(m))
  queryClient.removeQueries({ queryKey: queryKeys.sandbox.session(scope) })
}

export function removeSandboxSessionQueriesByInstanceId(queryClient: QueryClient, sandboxId: string): void {
  const scopes = new Set<string>()
  for (const query of queryClient.getQueryCache().findAll({ queryKey: queryKeys.sandbox.all })) {
    if (query.queryKey.includes(sandboxId)) {
      scopes.add(query.queryKey[1] as string)
    }
  }
  scopes.forEach((s) => removeSandboxSessionQueries(queryClient, s))
}

export function useSandboxSession(options?: UseSandboxSessionOptions): UseSandboxSessionResult {
  const { scope, createParams, terminal = false, vnc = false, notify } = options ?? {}
  const notifyRef = useRef({ sandbox: true, terminal: true, vnc: true, ...notify })
  notifyRef.current = { sandbox: true, terminal: true, vnc: true, ...notify }
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const { sandboxApi, toolboxApi } = useApi()
  const queryClient = useQueryClient()

  const client = useMemo(() => {
    if (!user?.access_token || !selectedOrganization?.id) return null
    return new Daytona({
      jwtToken: user.access_token,
      apiUrl: import.meta.env.VITE_API_URL,
      organizationId: selectedOrganization.id,
    })
  }, [user?.access_token, selectedOrganization?.id])

  const createMutation = useMutation<Sandbox, Error, CreateSandboxParams | undefined>({
    mutationKey: ['create-sandbox', scope ?? 'default'],
    mutationFn: async (params) => {
      if (!client) throw new Error('Unable to create Daytona client: missing access token or organization ID.')
      return await client.create(params ?? createParams)
    },
    onSuccess: (newSandbox) => {
      if (scope) queryClient.setQueryData(queryKeys.sandbox.currentId(scope), newSandbox.id)
    },
    onError: (error) => {
      if (notifyRef.current.sandbox) {
        toast.error('Failed to create sandbox', {
          description: error.message,
          action: { label: 'Try again', onClick: () => createMutation.mutate(createParams) },
        })
      }
    },
  })

  const persistedSandboxId = scope ? queryClient.getQueryData<string>(queryKeys.sandbox.currentId(scope)) : undefined
  const sandboxId = createMutation.data?.id ?? persistedSandboxId ?? ''
  const resolvedScope = scope ?? sandboxId

  const sandboxQuery = useQuery<Sandbox>({
    queryKey: queryKeys.sandbox.instance(resolvedScope, sandboxId),
    queryFn: () => client?.get(sandboxId) ?? Promise.reject(new Error('Client not initialized')),
    enabled: !!resolvedScope && !!sandboxId && !!client,
  })

  const sandbox = sandboxQuery.data ?? createMutation.data ?? null

  const getPortPreviewUrl = useCallback(
    async (id: string, port: number) =>
      (await sandboxApi.getSignedPortPreviewUrl(id, port, selectedOrganization?.id)).data.url,
    [sandboxApi, selectedOrganization?.id],
  )

  const terminalQuery = useQuery<string, Error>({
    queryKey: queryKeys.sandbox.terminalUrl(resolvedScope, sandboxId),
    queryFn: () => getPortPreviewUrl(sandboxId, TERMINAL_PORT),
    enabled: terminal && !!sandboxId,
    staleTime: Infinity,
  })

  const vncToastId = `vnc-${resolvedScope}-${sandboxId}`
  const vncToastShownRef = useRef(false)

  const startVncMutation = useMutation<void, Error>({
    mutationFn: async () => {
      await toolboxApi.startComputerUseDeprecated(sandboxId, selectedOrganization?.id)
    },
    onMutate: () => {
      if (notifyRef.current.vnc) {
        vncToastShownRef.current = true
        toast.loading('Starting VNC desktop...', { id: vncToastId })
      }
    },
    onSuccess: () => {
      if (vncToastShownRef.current) {
        toast.loading('VNC desktop started, checking status...', { id: vncToastId })
      }
    },
  })

  const prevSandboxIdRef = useRef<string>('')
  useEffect(() => {
    if (prevSandboxIdRef.current && !sandboxQuery.data && !sandboxQuery.isFetching) {
      createMutation.reset()
      startVncMutation.reset()
      vncToastShownRef.current = false
      if (scope) removeSandboxSessionQueries(queryClient, scope)
    }
    prevSandboxIdRef.current = sandboxId
  }, [sandboxId, sandboxQuery.data, sandboxQuery.isFetching, createMutation, startVncMutation, queryClient, scope])

  const vncStatusQuery = useQuery<string, Error>({
    queryKey: queryKeys.sandbox.vncStatus(resolvedScope, sandboxId),
    queryFn: async () => {
      const {
        data: { status },
      } = await toolboxApi.getComputerUseStatusDeprecated(sandboxId, selectedOrganization?.id)
      if (status !== 'active') throw new Error(`VNC desktop not ready: ${status}`)
      return status
    },
    enabled: vnc && !!sandboxId && startVncMutation.isSuccess,
  })

  const vncReady = vncStatusQuery.data === 'active'

  const vncUrlQuery = useQuery<string, Error>({
    queryKey: queryKeys.sandbox.vncUrl(resolvedScope, sandboxId),
    queryFn: async () => await getPortPreviewUrl(sandboxId, VNC_PORT),
    enabled: vnc && !!sandboxId && vncReady,
    staleTime: Infinity,
  })

  useEffect(() => {
    if (!vncToastShownRef.current) return

    if (vncUrlQuery.data) {
      toast.success('VNC desktop is ready', { id: vncToastId })
      vncToastShownRef.current = false
    } else if (startVncMutation.error) {
      toast.error('Failed to start VNC desktop', { id: vncToastId, description: startVncMutation.error.message })
      vncToastShownRef.current = false
    } else if (vncStatusQuery.error) {
      toast.error('VNC desktop failed to become ready', { id: vncToastId, description: vncStatusQuery.error.message })
      vncToastShownRef.current = false
    }
  }, [vncToastId, vncUrlQuery.data, startVncMutation.error, vncStatusQuery.error])

  const createSandbox = useCallback(
    (params?: CreateSandboxParams) => createMutation.mutateAsync(params ?? createParams),
    [createMutation, createParams],
  )

  return {
    sandbox: {
      instance: sandbox,
      loading: createMutation.isPending || (!!sandboxId && sandboxQuery.isLoading),
      error: createMutation.error?.message ?? sandboxQuery.error?.message ?? null,
      create: createSandbox,
    },
    terminal: {
      url: terminalQuery.data ?? null,
      loading: terminalQuery.isLoading,
      error: terminalQuery.error?.message ?? null,
      refetch: terminalQuery.refetch,
    },
    vnc: {
      url: vncUrlQuery.data ?? null,
      loading: startVncMutation.isPending || vncStatusQuery.isLoading || (vncReady && vncUrlQuery.isLoading),
      error: startVncMutation.error?.message ?? vncStatusQuery.error?.message ?? vncUrlQuery.error?.message ?? null,
      start: () => startVncMutation.mutate(),
      refetch: () => {
        startVncMutation.reset()
        queryClient.removeQueries({ queryKey: queryKeys.sandbox.vncStatus(resolvedScope, sandboxId) })
        queryClient.removeQueries({ queryKey: queryKeys.sandbox.vncUrl(resolvedScope, sandboxId) })
        if (notifyRef.current.vnc) {
          vncToastShownRef.current = true
          toast.loading('Retrying VNC desktop...', { id: vncToastId })
        }
        startVncMutation.mutate()
      },
    },
  }
}
