/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { io, Socket } from 'socket.io-client'

/** Handler receives (eventName, rawData). */
export type EventHandler = (eventName: string, data: any) => void

/**
 * Extracts the resource ID from an event payload.
 *
 * Handles two payload shapes:
 *   - Wrapper: {sandbox: {id: ...}, ...} -> nested resource ID
 *   - Direct: {id: ...} -> top-level ID
 */
function extractIdFromEvent(data: any): string | undefined {
  if (!data || typeof data !== 'object') return undefined
  for (const key of ['sandbox', 'volume', 'snapshot', 'runner']) {
    const nested = data[key]
    if (nested && typeof nested === 'object' && typeof nested.id === 'string') {
      return nested.id
    }
  }
  if (typeof data.id === 'string') {
    return data.id
  }
  return undefined
}

/**
 * Manages a Socket.IO connection to the Daytona notification gateway
 * and dispatches resource events to per-resource handlers.
 */
export class EventDispatcher {
  private socket: Socket | undefined
  private _connected = false
  private _closed = false
  private _failed = false
  private _failError: string | null = null
  private listeners = new Map<string, Set<EventHandler>>()
  private registeredEvents = new Set<string>()
  private connectPromise: Promise<void> | null = null
  private ensureConnectPromise: Promise<void> | null = null
  private reconnectAttempts = 0
  private readonly maxReconnectAttempts = 10
  // Cleared on connect/error/disconnect to prevent late state mutation.
  private _connectTimer: ReturnType<typeof setTimeout> | null = null
  private disconnectTimer: ReturnType<typeof setTimeout> | null = null
  private disconnectGeneration = 0
  private static readonly DISCONNECT_DELAY_MS = 30_000

  constructor(
    private readonly apiUrl: string,
    private readonly token: string,
    private readonly organizationId?: string,
  ) {}

  /**
   * Idempotent: ensure a connection attempt is in progress or already established.
   *
   * Non-blocking. Fires-and-forgets a connect() call via a stored promise if not
   * already connected and no attempt is currently running.
   */
  ensureConnected(): void {
    // No-op after disconnect — prevents socket resurrection.
    if (this._closed) return
    if (this._connected) return
    if (this.connectPromise) return
    if (this.ensureConnectPromise) return

    this.ensureConnectPromise = this.connect()
      .catch(() => {
        // Callers check isConnected when they need it
      })
      .finally(() => {
        this.ensureConnectPromise = null
      })
  }

  /**
   * Establishes the Socket.IO connection. Resolves when connected.
   * Throws if the connection fails within the timeout.
   */
  async connect(timeoutMs = 5000): Promise<void> {
    if (this._closed) {
      return
    }

    if (this._connected && this.socket) {
      return
    }

    if (this.connectPromise) {
      return this.connectPromise
    }

    this.connectPromise = this.doConnect(timeoutMs)

    try {
      await this.connectPromise

      if (this.listeners.size === 0) {
        this.scheduleDelayedDisconnect()
      }
    } catch (error) {
      if (this.listeners.size === 0) {
        // Stop retries if nobody is listening.
        this.socket?.disconnect()
      }
      throw error
    } finally {
      this.connectPromise = null
    }
  }

  private doConnect(timeoutMs: number): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      if (this.socket) {
        this.socket.removeAllListeners()
        this.socket.disconnect()
        this.socket = undefined
      }

      // Strip /api suffix to get the origin
      const origin = this.apiUrl.replace(/\/api\/?$/, '')

      const query: Record<string, string> = {}
      if (this.organizationId) {
        query.organizationId = this.organizationId
      }

      this.socket = io(origin, {
        path: '/api/socket.io/',
        autoConnect: false,
        transports: ['websocket'],
        query,
        reconnection: true,
        reconnectionAttempts: this.maxReconnectAttempts,
        reconnectionDelay: 1000,
        reconnectionDelayMax: 30000,
      })

      this.socket.auth = { token: this.token }

      this.clearConnectTimer()
      this._connectTimer = setTimeout(() => {
        if (!this._connected) {
          this.socket?.disconnect()
          this.clearConnectTimer()
          this._failed = true
          this._failError = 'WebSocket connection timed out'
          reject(new Error(this._failError))
        }
      }, timeoutMs)

      if (typeof this._connectTimer.unref === 'function') {
        this._connectTimer.unref()
      }

      this.socket.on('connect', () => {
        this.clearConnectTimer()
        this._connected = true
        this._failed = false
        this._failError = null
        this.reconnectAttempts = 0

        // Unref all underlying handles so they don't prevent Node.js process exit.
        // The socket.io connection should not keep the process alive — it's a background
        // enhancement, not a critical resource.
        this.unrefAll()

        resolve()
      })

      this.socket.on('connect_error', (err) => {
        if (!this._connected) {
          this.clearConnectTimer()
          this._failed = true
          this._failError = `WebSocket connection failed: ${err.message}`
          reject(new Error(this._failError))
        }
      })

      this.socket.on('disconnect', (reason) => {
        this._connected = false
        if (reason === 'io server disconnect') {
          // Server initiated disconnect - try to reconnect
          this.socket?.connect()
        }
      })

      this.socket.io.on('reconnect', () => {
        this._connected = true
        this._failed = false
        this._failError = null
        this.reconnectAttempts = 0
      })

      this.socket.io.on('reconnect_attempt', () => {
        this.reconnectAttempts++
      })

      this.socket.io.on('reconnect_failed', () => {
        this._connected = false
        this._failed = true
        this._failError = `WebSocket reconnection failed after ${this.maxReconnectAttempts} attempts`
      })

      // Re-register any events that were added before the socket was created
      const pendingEvents = [...this.registeredEvents]
      this.registeredEvents.clear()
      this.registerEvents(pendingEvents)

      this.socket.connect()
    })
  }

  /**
   * Registers Socket.IO event handlers (idempotent -- each event is registered once).
   */
  private registerEvents(events: string[]): void {
    for (const eventName of events) {
      if (this.registeredEvents.has(eventName)) {
        continue
      }
      this.registeredEvents.add(eventName)

      // If socket isn't created yet, the event will be registered when connect() runs
      if (!this.socket) continue

      const handler = (data: any) => {
        const resourceId = extractIdFromEvent(data)
        if (resourceId) {
          this.dispatch(resourceId, eventName, data)
        }
      }

      this.socket.on(eventName, handler)
    }
  }

  /**
   * Registers a handler for events targeting a specific resource.
   * Returns an unsubscribe function.
   *
   * @param resourceId - The ID of the resource (e.g. sandbox ID).
   * @param handler - Callback receiving (eventName, rawData).
   * @param events - List of Socket.IO event names to listen for.
   */
  subscribe(resourceId: string, handler: EventHandler, events: string[]): () => void {
    // No-op after disconnect — prevents socket resurrection.
    if (this._closed) {
      return () => {
        return
      }
    }

    this.cancelDelayedDisconnect()
    this.disconnectGeneration++
    this.ensureConnected()

    if (!this.listeners.has(resourceId)) {
      this.listeners.set(resourceId, new Set())
    }
    this.listeners.get(resourceId)!.add(handler)

    // Register any new events with the Socket.IO client
    this.registerEvents(events)

    return () => {
      const handlers = this.listeners.get(resourceId)
      if (handlers) {
        handlers.delete(handler)
        if (handlers.size === 0) {
          this.listeners.delete(resourceId)
        }
      }

      // Schedule delayed disconnect when no resources are listening anymore
      if (this.listeners.size === 0) {
        this.scheduleDelayedDisconnect()
      }
    }
  }

  /** Whether the WebSocket is currently connected */
  get isConnected(): boolean {
    return this._connected
  }

  /** Whether the WebSocket has permanently failed (exhausted reconnection attempts) */
  get isFailed(): boolean {
    return this._failed
  }

  /** The error message if the connection has failed */
  get failError(): string | null {
    return this._failError
  }

  /** Disconnects and cleans up all resources */
  disconnect(): void {
    this._closed = true
    this.cancelDelayedDisconnect()
    this.clearConnectTimer()
    this.connectPromise = null
    this.ensureConnectPromise = null
    this.disconnectSocket()
  }

  private disconnectSocket(): void {
    if (this.socket) {
      this.socket.removeAllListeners()
      this.socket.disconnect()
      this.socket = undefined
    }
    this._connected = false
    this.listeners.clear()
    this.registeredEvents.clear()
  }

  /**
   * Unrefs the underlying TCP socket so the event subscriber doesn't prevent
   * Node.js process exit. This is a no-op in browser environments.
   * Only unrefs the raw socket — does not touch engine.io timers or internals.
   */
  private unrefAll(): void {
    try {
      const engine = (this.socket?.io as any)?.engine
      if (!engine) return

      // Unref the raw TCP socket underneath the WebSocket transport.
      // This tells Node.js not to keep the event loop alive solely for this connection.
      const rawSocket = engine?.transport?.ws?._socket
      if (rawSocket && typeof rawSocket.unref === 'function') {
        rawSocket.unref()
      }
    } catch {
      // Not critical - only affects process exit behavior
    }
  }

  private dispatch(resourceId: string, eventName: string, data: any): void {
    if (!resourceId) return
    const handlers = this.listeners.get(resourceId)
    if (handlers) {
      for (const handler of handlers) {
        try {
          handler(eventName, data)
        } catch {
          // Don't let a handler error break other handlers
        }
      }
    }
  }

  private cancelDelayedDisconnect(): void {
    if (this.disconnectTimer) {
      clearTimeout(this.disconnectTimer)
      this.disconnectTimer = null
    }
  }

  private clearConnectTimer(): void {
    if (this._connectTimer) {
      clearTimeout(this._connectTimer)
      this._connectTimer = null
    }
  }

  private scheduleDelayedDisconnect(): void {
    this.cancelDelayedDisconnect()
    const generation = this.disconnectGeneration

    this.disconnectTimer = setTimeout(() => {
      if (generation !== this.disconnectGeneration) {
        return
      }

      if (this.listeners.size === 0) {
        this.disconnectSocket()
      }
    }, EventDispatcher.DISCONNECT_DELAY_MS)

    if (typeof this.disconnectTimer.unref === 'function') {
      this.disconnectTimer.unref()
    }
  }
}
