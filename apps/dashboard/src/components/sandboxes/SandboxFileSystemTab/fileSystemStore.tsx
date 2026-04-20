/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createContext, type ReactNode, useContext, useRef } from 'react'
import { createStore, useStore, type StoreApi } from 'zustand'

import type { SandboxFileSystemNode } from './types'
import { ROOT_NODE } from './constants'

type FileSystemStoreState = {
  deleteTarget: SandboxFileSystemNode | null
  folderCreationParentPath: string | null
  isContentsOverlayMode: boolean
  isContentsOverlayOpen: boolean
  isSearchOpen: boolean
  lastOpenedNodePath: string
  newFolderName: string
  openDropdownPath: string | null
  searchLabelAvailableWidth: number
  searchLabelFont: string
  searchQuery: string
  selectedNode: SandboxFileSystemNode | null
}

type FileSystemStoreActions = {
  closeCreateFolder: () => void
  closeDeleteDialog: () => void
  clearSelectedNode: () => void
  openNode: (value: SandboxFileSystemNode) => void
  resetSearch: () => void
  setLastOpenedNodePath: (value: string) => void
  setDeleteTarget: (value: SandboxFileSystemNode | null) => void
  setFolderCreationParentPath: (value: string | null) => void
  setContentsOverlayMode: (value: boolean) => void
  setContentsOverlayOpen: (value: boolean) => void
  setNewFolderName: (value: string) => void
  setOpenDropdownPath: (value: string | null) => void
  setSearchLabelMeasurements: (payload: { availableWidth: number; font: string }) => void
  setSearchOpen: (value: boolean) => void
  setSearchQuery: (value: string) => void
  setSelectedNode: (value: SandboxFileSystemNode | null) => void
}

export type FileSystemStore = FileSystemStoreState & { actions: FileSystemStoreActions }

const createFileSystemStore = () =>
  createStore<FileSystemStore>((set, get) => ({
    deleteTarget: null,
    folderCreationParentPath: null,
    isContentsOverlayMode: false,
    isContentsOverlayOpen: false,
    isSearchOpen: false,
    lastOpenedNodePath: ROOT_NODE.path,
    newFolderName: '',
    openDropdownPath: null,
    searchLabelAvailableWidth: 0,
    searchLabelFont: '',
    searchQuery: '',
    selectedNode: null,
    actions: {
      closeCreateFolder: () =>
        set({
          folderCreationParentPath: null,
          newFolderName: '',
        }),
      closeDeleteDialog: () =>
        set({
          deleteTarget: null,
        }),
      clearSelectedNode: () =>
        set({
          selectedNode: null,
          isContentsOverlayOpen: false,
          openDropdownPath: null,
        }),
      openNode: (selectedNode) =>
        set({
          selectedNode,
          lastOpenedNodePath: selectedNode.path,
          isContentsOverlayOpen: get().isContentsOverlayMode ? true : get().isContentsOverlayOpen,
        }),
      resetSearch: () =>
        set({
          searchQuery: '',
        }),
      setLastOpenedNodePath: (lastOpenedNodePath) => set({ lastOpenedNodePath }),
      setDeleteTarget: (deleteTarget) => set({ deleteTarget }),
      setFolderCreationParentPath: (folderCreationParentPath) => set({ folderCreationParentPath }),
      setContentsOverlayMode: (isContentsOverlayMode) => set({ isContentsOverlayMode }),
      setContentsOverlayOpen: (isContentsOverlayOpen) => set({ isContentsOverlayOpen }),
      setNewFolderName: (newFolderName) => set({ newFolderName }),
      setOpenDropdownPath: (openDropdownPath) => set({ openDropdownPath }),
      setSearchLabelMeasurements: ({ availableWidth, font }) =>
        set({
          searchLabelAvailableWidth: availableWidth,
          searchLabelFont: font,
        }),
      setSearchOpen: (isSearchOpen) => set({ isSearchOpen }),
      setSearchQuery: (searchQuery) => set({ searchQuery }),
      setSelectedNode: (selectedNode) => set({ selectedNode }),
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
