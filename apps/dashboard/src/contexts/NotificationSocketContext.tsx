/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { createContext } from 'react'
import { Socket } from 'socket.io-client'

export interface INotificationSocketContext {
  notificationSocket: Socket | null
}

export const NotificationSocketContext = createContext<INotificationSocketContext | undefined>(undefined)
