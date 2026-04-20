/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Buffer } from 'buffer'
import { keepPreviousData, useQuery, type QueryClient, type UseQueryOptions } from '@tanstack/react-query'

import { handleApiError } from '@/lib/error-handling'

import type { PreviewKind, SandboxFileSystemNode, SandboxInstance } from './types'
import { ROOT_PATH } from './constants'
import { getImageMimeType, isProbablyBinary, sortEntries, toNode } from './utils'

export type FilePreviewData = {
  content?: string
  imageUrl?: string
  kind: PreviewKind
}

export const fileSystemQueryKeys = {
  all: (sandboxId: string) => ['sandbox-file-system', sandboxId] as const,
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
  path,
  sandboxInstance,
}: {
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
        handleApiError(error, `Failed to list ${path}`)
        throw error
      }
    },
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

export function useFilePreviewQuery({
  enabled,
  path,
  sandboxInstance,
}: {
  enabled: boolean
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
            imageUrl: `data:${imageMimeType};base64,${fileContents.toString('base64')}`,
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
        handleApiError(error, `Failed to read ${path}`)
        throw error
      }
    },
    enabled,
    placeholderData: keepPreviousData,
    staleTime: 0,
  })
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
        handleApiError(error, `Failed to search ${ROOT_PATH}`)
        throw error
      }
    },
    enabled,
    placeholderData: keepPreviousData,
    staleTime: 30_000,
  })
}
