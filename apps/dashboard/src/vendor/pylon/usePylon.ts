/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useConfig } from '@/hooks/useConfig'
import { useEffect } from 'react'
import { create } from 'zustand'

interface PylonState {
  isOpen: boolean
  unreadCount: number
  isInitialized: boolean
  initListeners: () => void
  toggle: () => void
}

const usePylonStore = create<PylonState>()((set, get) => ({
  isOpen: false,
  unreadCount: 0,
  isInitialized: false,

  initListeners: () => {
    if (get().isInitialized || !window.Pylon) {
      return
    }

    window.Pylon('onShow', () => set({ isOpen: true }))
    window.Pylon('onHide', () => set({ isOpen: false }))
    window.Pylon('onChangeUnreadMessagesCount', (count: number) => set({ unreadCount: count }))

    set({ isInitialized: true })
  },

  toggle: () => {
    if (!window.Pylon) return

    if (get().isOpen) {
      window.Pylon('hide')
    } else {
      window.Pylon('show')
    }
  },
}))

export function usePylon() {
  const config = useConfig()
  const pylon = usePylonStore()

  const initListeners = usePylonStore((state) => state.initListeners)

  useEffect(() => {
    initListeners()
  }, [initListeners])

  return {
    isOpen: pylon.isOpen,
    unreadCount: pylon.unreadCount,
    toggle: pylon.toggle,
    isEnabled: Boolean(config.pylonAppId),
  }
}
