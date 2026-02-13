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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

type CreateSandboxParams = CreateSandboxBaseParams | CreateSandboxFromImageParams | CreateSandboxFromSnapshotParams

const TERMINAL_PORT = 22222
const VNC_PORT = 6080

export type UseSandboxSessionOptions = {
  key?: string
  createParams?: CreateSandboxParams
  autoCreate?: boolean
  terminal?: boolean
  vnc?: boolean
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

export function useSandboxSession(options?: UseSandboxSessionOptions): UseSandboxSessionResult {
  const { key: sessionKey, autoCreate = false, createParams, terminal = false, vnc = false } = options ?? {}
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
    mutationFn: async (params) => {
      if (!client) throw new Error('Unable to create Daytona client: missing access token or organization ID.')
      return await client.create(params ?? createParams)
    },
    onSuccess: (newSandbox) => {
      queryClient.setQueryData(queryKeys.sandbox.instance(sessionKey ?? newSandbox.id), newSandbox)
    },
    onError: (error) => {
      toast.error('Failed to create sandbox', {
        description: error.message,
        action: { label: 'Try again', onClick: () => createMutation.mutate(createParams) },
      })
    },
  })

  // Derive sandbox ID from cache first, then from mutation result
  const cachedSandbox = sessionKey
    ? queryClient.getQueryData<Sandbox>(queryKeys.sandbox.instance(sessionKey))
    : undefined
  const sandboxId = cachedSandbox?.id ?? createMutation.data?.id ?? ''
  const key = sessionKey ?? createMutation.data?.id ?? ''

  const sandboxQuery = useQuery<Sandbox>({
    queryKey: queryKeys.sandbox.instance(key),
    queryFn: () => client?.get(sandboxId) ?? Promise.reject(new Error('Client not initialized')),
    enabled: !!key && !!sandboxId && !!client,
    staleTime: Infinity,
  })

  const sandbox = sandboxQuery.data ?? null

  useEffect(() => {
    if (autoCreate && createMutation.status === 'idle' && !cachedSandbox) {
      createMutation.mutate(createParams)
    }
  }, [autoCreate, createMutation.status, cachedSandbox, createParams])

  const getPortPreviewUrl = useCallback(
    async (sandboxId: string, port: number) =>
      (await sandboxApi.getSignedPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url,
    [sandboxApi, selectedOrganization?.id],
  )

  const terminalQuery = useQuery<string, Error>({
    queryKey: queryKeys.sandbox.terminalUrl(key),
    queryFn: () => getPortPreviewUrl(sandboxId, TERMINAL_PORT),
    enabled: terminal && !!sandboxId,
    staleTime: Infinity,
  })

  const vncToastId = `vnc-${key}`

  const startVncMutation = useMutation<void, Error>({
    mutationFn: async () => {
      await toolboxApi.startComputerUseDeprecated(sandboxId, selectedOrganization?.id)
    },
    onMutate: () => {
      toast.loading('Starting VNC desktop...', { id: vncToastId })
    },
    onSuccess: () => {
      toast.loading('VNC desktop started, checking status...', { id: vncToastId })
    },
    onError: (error) => {
      toast.error('Failed to start VNC desktop', { id: vncToastId, description: error.message })
    },
  })

  const vncStatusQuery = useQuery<string, Error>({
    queryKey: queryKeys.sandbox.vncStatus(key),
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
    queryKey: queryKeys.sandbox.vncUrl(key),
    queryFn: async () => (await getPortPreviewUrl(sandboxId, VNC_PORT)) + '/vnc.html?autoconnect=true',
    enabled: vnc && !!sandboxId && vncStatusQuery.data === 'active',
    staleTime: Infinity,
  })

  useEffect(() => {
    if (vncStatusQuery.error) {
      toast.error('VNC desktop failed to become ready', { id: vncToastId, description: vncStatusQuery.error.message })
    }
  }, [vncStatusQuery.error, vncToastId])

  useEffect(() => {
    if (vncUrlQuery.data) {
      toast.success('VNC desktop is ready', { id: vncToastId })
    }
  }, [vncUrlQuery.data, vncToastId])

  const createSandbox = useCallback(
    (params?: CreateSandboxParams) => createMutation.mutateAsync(params ?? createParams),
    [createMutation, createParams],
  )

  return {
    sandbox: {
      instance: sandbox,
      loading: createMutation.isPending || (autoCreate && !sandbox && !createMutation.error),
      error: createMutation.error?.message ?? null,
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
        queryClient.removeQueries({ queryKey: queryKeys.sandbox.vncStatus(key) })
        queryClient.removeQueries({ queryKey: queryKeys.sandbox.vncUrl(key) })
        toast.loading('Retrying VNC desktop...', { id: vncToastId })
        startVncMutation.mutate()
      },
    },
  }
}
