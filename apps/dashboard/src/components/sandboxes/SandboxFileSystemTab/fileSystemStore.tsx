/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createContext, type ReactNode, useContext, useRef } from 'react'
import { createStore, useStore, type StoreApi } from 'zustand'

import { ROOT_NODE } from './constants'

type FileSystemStoreState = {
  lastOpenedNodePath: string
  nextFilePath: string | null
  previousFilePath: string | null
  selectedNodePath: string | null
}

type FileSystemStoreActions = {
  clearSelectedNode: () => void
  openNode: (path: string) => void
  setAdjacentFilePaths: (value: { nextFilePath: string | null; previousFilePath: string | null }) => void
  setSelectedNodePath: (value: string | null) => void
}

export type FileSystemStore = FileSystemStoreState & { actions: FileSystemStoreActions }

const createFileSystemStore = () =>
  createStore<FileSystemStore>((set) => ({
    lastOpenedNodePath: ROOT_NODE.path,
    nextFilePath: null,
    previousFilePath: null,
    selectedNodePath: null,
    actions: {
      clearSelectedNode: () =>
        set({
          nextFilePath: null,
          previousFilePath: null,
          selectedNodePath: null,
        }),
      openNode: (selectedNodePath) =>
        set({
          selectedNodePath,
          lastOpenedNodePath: selectedNodePath,
        }),
      setAdjacentFilePaths: ({ nextFilePath, previousFilePath }) =>
        set({
          nextFilePath,
          previousFilePath,
        }),
      setSelectedNodePath: (selectedNodePath) => set({ selectedNodePath }),
    },
  }))

const FileSystemStoreContext = createContext<StoreApi<FileSystemStore> | null>(null)

export function FileSystemStoreProvider({ children }: { children: ReactNode }) {
  const storeRef = useRef<StoreApi<FileSystemStore> | null>(null)

  if (!storeRef.current) {
    storeRef.current = createFileSystemStore()
  }

  return <FileSystemStoreContext value={storeRef.current}>{children}</FileSystemStoreContext>
}

export function useFileSystemStore<T>(selector: (state: FileSystemStore) => T): T {
  const store = useContext(FileSystemStoreContext)

  if (!store) {
    throw new Error('useFileSystemStore must be used within <FileSystemStoreProvider />')
  }

  return useStore(store, selector)
}
