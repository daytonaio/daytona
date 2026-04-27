/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Buffer } from 'buffer'
import { useCallback, useMemo } from 'react'
import {
  keepPreviousData,
  useIsFetching,
  useQueries,
  useQuery,
  useQueryClient,
  type QueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query'

import type { PreviewKind, SandboxFileSystemNode, SandboxInstance } from './types'
import { ROOT_NODE, ROOT_PATH } from './constants'
import {
  getImageMimeType,
  getParentPath,
  handleFileSystemApiError,
  isProbablyBinary,
  shouldRetryFileSystemQuery,
  sortEntries,
  toNode,
} from './utils'

export type FilePreviewData = {
  content?: string
  imageBlob?: Blob
  kind: PreviewKind
}

const FILE_PREVIEW_STALE_TIME = 60_000

export const fileSystemQueryKeys = {
  all: (sandboxId: string) => ['sandbox-file-system', sandboxId] as const,
  details: (sandboxId: string, path: string) => [...fileSystemQueryKeys.all(sandboxId), 'details', path] as const,
  directory: (sandboxId: string, path: string) => [...fileSystemQueryKeys.all(sandboxId), 'directory', path] as const,
  preview: (sandboxId: string, path: string) => [...fileSystemQueryKeys.all(sandboxId), 'preview', path] as const,
  search: (sandboxId: string, query: string) => [...fileSystemQueryKeys.all(sandboxId), 'search', query] as const,
  searchPrefix: (sandboxId: string) => [...fileSystemQueryKeys.all(sandboxId), 'search'] as const,
  sandbox: (sandboxId: string) => [...fileSystemQueryKeys.all(sandboxId), 'sandbox'] as const,
}

export function useSandboxInstanceQuery({
  client,
  sandboxId,
}: {
  client: { get: (sandboxId: string) => Promise<SandboxInstance> } | null
  sandboxId: string
}) {
  return useQuery({
    queryKey: fileSystemQueryKeys.sandbox(sandboxId),
    queryFn: () => {
      if (!client) {
        throw new Error('Unable to initialize Daytona client')
      }

      return client.get(sandboxId)
    },
    enabled: Boolean(client),
    staleTime: Number.POSITIVE_INFINITY,
  })
}

export function getDirectoryChildrenQueryOptions({
  notifyOnError = false,
  path,
  sandboxInstance,
}: {
  notifyOnError?: boolean
  path: string
  sandboxInstance: SandboxInstance
}) {
  return {
    queryKey: fileSystemQueryKeys.directory(sandboxInstance.id, path),
    queryFn: async () => {
      try {
        const files = sortEntries(await sandboxInstance.fs.listFiles(path))
        return files.map((file) => toNode(path, file))
      } catch (error) {
        if (notifyOnError) {
          handleFileSystemApiError(error, `Failed to list ${path}`, {
            toastId: `filesystem-list-${sandboxInstance.id}-${path}`,
          })
        }
        throw error
      }
    },
    placeholderData: undefined,
    retry: shouldRetryFileSystemQuery,
    staleTime: 60_000,
  } satisfies UseQueryOptions<SandboxFileSystemNode[]>
}

export async function invalidateDirectoryQuery({
  path,
  queryClient,
  sandboxInstance,
}: {
  path: string
  queryClient: QueryClient
  sandboxInstance: SandboxInstance
}) {
  await queryClient.invalidateQueries({
    queryKey: fileSystemQueryKeys.directory(sandboxInstance.id, path),
  })
}

export function useInvalidateDirectoryQuery({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const queryClient = useQueryClient()

  return useCallback(
    async (path: string) => {
      if (!sandboxInstance) {
        return
      }

      await queryClient.invalidateQueries({
        queryKey: fileSystemQueryKeys.directory(sandboxInstance.id, path),
      })
    },
    [queryClient, sandboxInstance],
  )
}

export function getFileDetailsQueryOptions({
  notifyOnError = false,
  path,
  sandboxInstance,
}: {
  notifyOnError?: boolean
  path: string
  sandboxInstance: SandboxInstance
}) {
  return {
    queryKey: fileSystemQueryKeys.details(sandboxInstance.id, path),
    queryFn: async (): Promise<SandboxFileSystemNode> => {
      try {
        if (path === ROOT_PATH) {
          return ROOT_NODE
        }

        const file = await sandboxInstance.fs.getFileDetails(path)
        return {
          ...toNode(getParentPath(path), file),
          id: path,
          path,
        }
      } catch (error) {
        if (notifyOnError) {
          handleFileSystemApiError(error, `Failed to load details for ${path}`, {
            toastId: `filesystem-details-${sandboxInstance.id}-${path}`,
          })
        }
        throw error
      }
    },
    placeholderData: undefined,
    retry: shouldRetryFileSystemQuery,
    staleTime: 60_000,
  } satisfies UseQueryOptions<SandboxFileSystemNode>
}

export async function invalidateFileDetailsQuery({
  path,
  queryClient,
  sandboxInstance,
}: {
  path: string
  queryClient: QueryClient
  sandboxInstance: SandboxInstance
}) {
  await queryClient.invalidateQueries({
    queryKey: fileSystemQueryKeys.details(sandboxInstance.id, path),
  })
}

export async function invalidateFilePreviewQuery({
  path,
  queryClient,
  sandboxInstance,
}: {
  path: string
  queryClient: QueryClient
  sandboxInstance: SandboxInstance
}) {
  await queryClient.invalidateQueries({
    queryKey: fileSystemQueryKeys.preview(sandboxInstance.id, path),
  })
}

export function useFileDetailsQuery({
  enabled,
  path,
  sandboxInstance,
}: {
  enabled: boolean
  path: string
  sandboxInstance: SandboxInstance | undefined
}) {
  return useQuery({
    ...(sandboxInstance
      ? getFileDetailsQueryOptions({
          path,
          sandboxInstance,
        })
      : {
          queryKey: fileSystemQueryKeys.details('unknown', path),
          queryFn: async (): Promise<SandboxFileSystemNode> => {
            throw new Error('Sandbox instance is not available')
          },
        }),
    enabled,
  })
}

export function useFetchFileDetailsQuery({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const queryClient = useQueryClient()

  return useCallback(
    async (path: string) => {
      if (!sandboxInstance) {
        throw new Error('Sandbox instance is not available')
      }

      return queryClient.fetchQuery(
        getFileDetailsQueryOptions({
          path,
          sandboxInstance,
        }),
      )
    },
    [queryClient, sandboxInstance],
  )
}

export function useFetchDirectoryChildrenQuery({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const queryClient = useQueryClient()

  return useCallback(
    async (path: string) => {
      if (!sandboxInstance) {
        throw new Error('Sandbox instance is not available')
      }

      const children = await queryClient.fetchQuery(
        getDirectoryChildrenQueryOptions({
          path,
          sandboxInstance,
        }),
      )

      children.forEach((node) => {
        queryClient.setQueryData(fileSystemQueryKeys.details(sandboxInstance.id, node.path), node)
      })

      return children
    },
    [queryClient, sandboxInstance],
  )
}

export function useFileDetailsCache({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const queryClient = useQueryClient()

  const getCachedNode = useCallback(
    (path: string) => {
      if (path === ROOT_PATH) {
        return ROOT_NODE
      }

      if (!sandboxInstance) {
        return undefined
      }

      return queryClient.getQueryData<SandboxFileSystemNode>(fileSystemQueryKeys.details(sandboxInstance.id, path))
    },
    [queryClient, sandboxInstance],
  )

  return {
    getCachedNode,
  }
}

export function useFileDetailsQueries({
  paths,
  sandboxInstance,
}: {
  paths: string[]
  sandboxInstance: SandboxInstance | undefined
}) {
  const detailQueries = useQueries({
    queries: sandboxInstance
      ? paths.map((path) => ({
          ...getFileDetailsQueryOptions({
            path,
            sandboxInstance,
          }),
          enabled: false,
        }))
      : [],
  })

  const fileDetailsByPath = useMemo(() => {
    return new Map(
      paths.flatMap((path, index) => {
        const node = detailQueries[index]?.data
        return node ? [[path, node] as const] : []
      }),
    )
  }, [detailQueries, paths])

  return fileDetailsByPath
}

export function useIsDirectoryRefreshing({
  path,
  sandboxInstance,
}: {
  path: string | null
  sandboxInstance: SandboxInstance | undefined
}) {
  const fetchCount = useIsFetching({
    queryKey: sandboxInstance && path ? fileSystemQueryKeys.directory(sandboxInstance.id, path) : undefined,
  })

  return Boolean(sandboxInstance && path) && fetchCount > 0
}

export function useFilePreviewQuery({
  enabled,
  notifyOnError = false,
  path,
  sandboxInstance,
}: {
  enabled: boolean
  notifyOnError?: boolean
  path: string
  sandboxInstance: SandboxInstance | undefined
}) {
  return useQuery({
    queryKey: fileSystemQueryKeys.preview(sandboxInstance?.id ?? 'unknown', path),
    queryFn: async (): Promise<FilePreviewData> => {
      try {
        if (!sandboxInstance) {
          throw new Error('Sandbox instance is not available')
        }

        const fileContents = Buffer.from(await sandboxInstance.fs.downloadFile(path))
        const imageMimeType = getImageMimeType(path)

        if (imageMimeType) {
          return {
            imageBlob: new Blob([fileContents], { type: imageMimeType }),
            kind: 'image',
          }
        }

        if (isProbablyBinary(fileContents)) {
          return { kind: 'binary' }
        }

        return {
          content: fileContents.toString('utf8'),
          kind: 'text',
        }
      } catch (error) {
        if (notifyOnError) {
          handleFileSystemApiError(error, `Failed to read ${path}`, {
            toastId: `filesystem-preview-${sandboxInstance?.id ?? 'unknown'}-${path}`,
          })
        }
        throw error
      }
    },
    enabled,
    placeholderData: undefined,
    staleTime: FILE_PREVIEW_STALE_TIME,
  })
}

export function useIsFilePreviewRefreshing({
  path,
  sandboxInstance,
}: {
  path: string | null
  sandboxInstance: SandboxInstance | undefined
}) {
  const fetchCount = useIsFetching({
    queryKey: sandboxInstance && path ? fileSystemQueryKeys.preview(sandboxInstance.id, path) : undefined,
  })

  return Boolean(sandboxInstance && path) && fetchCount > 0
}

export function useFileSearchQuery({
  enabled,
  query,
  sandboxInstance,
}: {
  enabled: boolean
  query: string
  sandboxInstance: SandboxInstance | undefined
}) {
  return useQuery({
    queryKey: fileSystemQueryKeys.search(sandboxInstance?.id ?? 'unknown', query),
    queryFn: async (): Promise<string[]> => {
      try {
        if (!sandboxInstance) {
          throw new Error('Sandbox instance is not available')
        }

        const response = await sandboxInstance.fs.searchFiles(ROOT_PATH, `*${query}*`)
        return response.files
      } catch (error) {
        handleFileSystemApiError(error, `Failed to search ${ROOT_PATH}`, {
          toastId: `filesystem-search-${sandboxInstance?.id ?? 'unknown'}-${query}`,
        })
        throw error
      }
    },
    enabled: enabled && Boolean(sandboxInstance),
    placeholderData: keepPreviousData,
    staleTime: 30_000,
  })
}
