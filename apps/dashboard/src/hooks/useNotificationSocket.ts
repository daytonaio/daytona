/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotificationSocketContext } from '@/contexts/NotificationSocketContext'
import { useContext } from 'react'

export function useNotificationSocket() {
  const context = useContext(NotificationSocketContext)

  if (!context) {
    throw new Error('useNotificationSocket must be used within a NotificationSocketProvider')
  }

  return context
}
