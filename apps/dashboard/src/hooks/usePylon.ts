/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCallback, useEffect, useState } from 'react'

export function usePylon() {
  const [isOpen, setIsOpen] = useState(false)
  const [unreadCount, setUnreadCount] = useState(0)

  useEffect(() => {
    if (!window.Pylon) return
    window.Pylon('onShow', () => setIsOpen(true))
    window.Pylon('onHide', () => setIsOpen(false))
    window.Pylon('onChangeUnreadMessagesCount', (count) => setUnreadCount(count))
    return () => {
      if (!window.Pylon) return
      window.Pylon('onShow', null)
      window.Pylon('onHide', null)
      window.Pylon('onChangeUnreadMessagesCount', null)
    }
  }, [])

  const toggle = useCallback(() => {
    if (!window.Pylon) return
    if (isOpen) {
      window.Pylon('hide')
    } else {
      window.Pylon('show')
    }
  }, [isOpen])

  return { isOpen, unreadCount, toggle }
}
