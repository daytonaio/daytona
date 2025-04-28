import { createContext } from 'react'
import { Socket } from 'socket.io-client'

export interface INotificationSocketContext {
  notificationSocket: Socket
}

export const NotificationSocketContext = createContext<INotificationSocketContext | undefined>(undefined)
