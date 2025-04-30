/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type SortingState = {
  [key: string]: {
    field: string
    direction: 'asc' | 'desc'
  }
}

interface DashboardStore {
  sortingStates: SortingState
  updateSortingState: (viewId: string, field: string, direction: 'asc' | 'desc') => void
}

export const useDashboardStore = create<DashboardStore>()(
  persist(
    (set) => ({
      sortingStates: {},
      updateSortingState: (viewId, field, direction) =>
        set((state) => ({
          sortingStates: {
            ...state.sortingStates,
            [viewId]: { field, direction },
          },
        })),
    }),
    {
      name: 'dashboard-sorting-storage', // unique name for localStorage
    },
  ),
)
