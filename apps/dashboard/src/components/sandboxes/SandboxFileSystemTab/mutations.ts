/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Buffer } from 'buffer'
import { useMutation, useQueryClient } from '@tanstack/react-query'

import type { SandboxFileSystemNode, SandboxInstance } from './types'
import { ROOT_PATH } from './constants'
import { getParentPath, isSameOrDescendantPath } from './utils'
import { fileSystemQueryKeys, invalidateDirectoryQuery, invalidateFileDetailsQuery } from './queries'

export function useCreateFolderMutation({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ path, permissions }: { path: string; permissions: string }) => {
      if (!sandboxInstance) {
        throw new Error('Sandbox instance is not available')
      }

      await sandboxInstance.fs.createFolder(path, permissions)
      return { path, permissions }
    },
    onSuccess: async ({ path }) => {
      if (!sandboxInstance) {
        return
      }

      await invalidateDirectoryQuery({
        path: getParentPath(path),
        queryClient,
        sandboxInstance,
      })

      await invalidateFileDetailsQuery({
        path,
        queryClient,
        sandboxInstance,
      })
    },
  })
}

export function useDeleteNodeMutation({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (node: SandboxFileSystemNode) => {
      if (!sandboxInstance) {
        throw new Error('Sandbox instance is not available')
      }

      await sandboxInstance.fs.deleteFile(node.path, node.isDir)
      return node
    },
    onSuccess: async (node) => {
      if (!sandboxInstance) {
        return
      }

      await invalidateDirectoryQuery({
        path: getParentPath(node.path),
        queryClient,
        sandboxInstance,
      })

      queryClient.removeQueries({
        queryKey: fileSystemQueryKeys.details(sandboxInstance.id, node.path),
      })

      queryClient.removeQueries({
        queryKey: fileSystemQueryKeys.preview(sandboxInstance.id, node.path),
      })

      queryClient.setQueriesData<string[]>(
        {
          queryKey: fileSystemQueryKeys.searchPrefix(sandboxInstance.id),
        },
        (currentResults) => currentResults?.filter((result) => !isSameOrDescendantPath(result, node.path)),
      )
    },
  })
}

export function useMoveNodeMutation({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ destinationPath, node }: { destinationPath: string; node: SandboxFileSystemNode }) => {
      if (!sandboxInstance) {
        throw new Error('Sandbox instance is not available')
      }

      await sandboxInstance.fs.moveFiles(node.path, destinationPath)

      return {
        destinationParentPath: getParentPath(destinationPath),
        destinationPath,
        node,
        sourceParentPath: getParentPath(node.path),
      }
    },
    onSuccess: async ({ destinationParentPath, destinationPath, node, sourceParentPath }) => {
      if (!sandboxInstance) {
        return
      }

      await invalidateDirectoryQuery({
        path: sourceParentPath,
        queryClient,
        sandboxInstance,
      })

      if (destinationParentPath !== sourceParentPath) {
        await invalidateDirectoryQuery({
          path: destinationParentPath,
          queryClient,
          sandboxInstance,
        })
      }

      queryClient.removeQueries({
        predicate: (query) => {
          const [scope, sandboxId, kind, path] = query.queryKey

          return (
            scope === 'sandbox-file-system' &&
            sandboxId === sandboxInstance.id &&
            (kind === 'details' || kind === 'directory' || kind === 'preview') &&
            typeof path === 'string' &&
            isSameOrDescendantPath(path, node.path)
          )
        },
      })

      await invalidateFileDetailsQuery({
        path: destinationPath,
        queryClient,
        sandboxInstance,
      })

      await queryClient.invalidateQueries({
        queryKey: fileSystemQueryKeys.searchPrefix(sandboxInstance.id),
      })
    },
  })
}

export function useUploadFilesMutation({ sandboxInstance }: { sandboxInstance: SandboxInstance | undefined }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ files, targetPath }: { files: File[]; targetPath: string }) => {
      if (!sandboxInstance) {
        throw new Error('Sandbox instance is not available')
      }

      const uploads = await Promise.all(
        files.map(async (file) => ({
          source: Buffer.from(await file.arrayBuffer()),
          destination: targetPath === ROOT_PATH ? `/${file.name}` : `${targetPath}/${file.name}`,
        })),
      )

      await sandboxInstance.fs.uploadFiles(uploads)

      return {
        files,
        targetPath,
      }
    },
    onSuccess: async ({ targetPath }) => {
      if (!sandboxInstance) {
        return
      }

      await invalidateDirectoryQuery({
        path: targetPath,
        queryClient,
        sandboxInstance,
      })

      await queryClient.invalidateQueries({
        queryKey: fileSystemQueryKeys.searchPrefix(sandboxInstance.id),
      })
    },
  })
}
