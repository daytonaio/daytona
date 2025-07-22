/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useAuth } from 'react-oidc-context'
import { ReactNode, useEffect, useRef } from 'react'
import { io } from 'socket.io-client'
import { NotificationSocketContext } from '@/contexts/NotificationSocketContext'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'

type Props = {
  children: ReactNode
}

export function NotificationSocketProvider(props: Props) {
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()

  const notificationSocketRef = useRef<ReturnType<typeof io> | null>(null)

  useEffect(() => {
    notificationSocketRef.current = io(window.location.origin, {
      path: '/api/socket.io/',
      autoConnect: false,
      transports: ['websocket', 'webtransport'],
      query: {
        organizationId: selectedOrganization?.id,
      },
    })

    const socket = notificationSocketRef.current
    if (user) {
      const token = user.access_token
      socket.auth = { token }
      socket.connect()
    }

    return () => {
      socket.disconnect()
    }
  }, [user, selectedOrganization?.id])

  return (
    <NotificationSocketContext.Provider value={{ notificationSocket: notificationSocketRef.current }}>
      {props.children}
    </NotificationSocketContext.Provider>
  )
}
