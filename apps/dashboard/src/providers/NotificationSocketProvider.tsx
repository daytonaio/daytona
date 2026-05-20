/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NotificationSocketContext } from '@/contexts/NotificationSocketContext'
import { useConfig } from '@/hooks/useConfig'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { ReactNode, useEffect, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { io, Socket } from 'socket.io-client'

type Props = {
  children: ReactNode
}

function getNotificationSocketUrl(apiUrl: string) {
  const url = new URL(apiUrl)
  const apiPath = url.pathname.replace(/\/$/, '') || '/api'

  return {
    origin: url.origin,
    path: `${apiPath}/socket.io/`,
  }
}

export function NotificationSocketProvider(props: Props) {
  const { user } = useAuth()
  const { apiUrl } = useConfig()
  const { selectedOrganization } = useSelectedOrganization()
  const [notificationSocket, setNotificationSocket] = useState<Socket | null>(null)

  useEffect(() => {
    const { origin, path } = getNotificationSocketUrl(apiUrl)
    const socket = io(origin, {
      path,
      autoConnect: false,
      transports: ['websocket', 'webtransport'],
      query: {
        organizationId: selectedOrganization?.id,
      },
    })

    setNotificationSocket(socket)

    if (user) {
      const token = user.access_token
      socket.auth = { token }
      socket.connect()
    }

    return () => {
      socket.disconnect()
    }
  }, [user, selectedOrganization?.id, apiUrl])

  return (
    <NotificationSocketContext.Provider value={{ notificationSocket }}>
      {props.children}
    </NotificationSocketContext.Provider>
  )
}
