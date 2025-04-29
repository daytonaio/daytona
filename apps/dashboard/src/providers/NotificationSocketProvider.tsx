/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useAuth } from 'react-oidc-context'
import { ReactNode, useEffect, useRef } from 'react'
import { io } from 'socket.io-client'
import { NotificationSocketContext } from '@/contexts/NotificationSocketContext'

type Props = {
  children: ReactNode
}

export function NotificationSocketProvider(props: Props) {
  const { user } = useAuth()
  const notificationSocketRef = useRef(
    io(window.location.origin, {
      path: '/api/socket.io/',
      autoConnect: false,
      transports: ['websocket', 'webtransport'],
    }),
  )

  useEffect(() => {
    const socket = notificationSocketRef.current
    if (user) {
      const token = user.access_token
      socket.auth = { token }
      socket.connect()
    }
    return () => {
      socket.disconnect()
    }
  }, [user])

  return (
    <NotificationSocketContext.Provider value={{ notificationSocket: notificationSocketRef.current }}>
      {props.children}
    </NotificationSocketContext.Provider>
  )
}
