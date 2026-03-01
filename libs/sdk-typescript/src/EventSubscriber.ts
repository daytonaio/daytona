/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { io, Socket } from 'socket.io-client'
import { Sandbox as SandboxDto, SandboxState, SandboxDesiredState } from '@daytonaio/api-client'

/** Event payload for sandbox.state.updated */
export interface SandboxStateUpdatedEvent {
  sandbox: SandboxDto
  oldState: SandboxState
  newState: SandboxState
}

/** Event payload for sandbox.desired-state.updated */
export interface SandboxDesiredStateUpdatedEvent {
  sandbox: SandboxDto
  oldDesiredState: SandboxDesiredState
  newDesiredState: SandboxDesiredState
}

/** Union of all sandbox event payloads dispatched to handlers */
export type SandboxEvent =
  | { type: 'state.updated'; data: SandboxStateUpdatedEvent }
  | { type: 'desired-state.updated'; data: SandboxDesiredStateUpdatedEvent }
  | { type: 'created'; data: SandboxDto }

/** Handler function for sandbox events */
export type SandboxEventHandler = (event: SandboxEvent) => void

/**
 * Manages a Socket.IO connection to the Daytona notification gateway
 * and dispatches sandbox events to per-sandbox handlers.
 */
export class EventSubscriber {
  private socket: Socket | null = null
  private _connected = false
  private _failed = false
  private _failError: string | null = null
  private listeners = new Map<string, Set<SandboxEventHandler>>()
  private connectPromise: Promise<void> | null = null
  private reconnectAttempts = 0
  private readonly maxReconnectAttempts = 10
  private disconnectTimer: ReturnType<typeof setTimeout> | null = null
  private static readonly DISCONNECT_DELAY_MS = 30_000

  constructor(
    private readonly apiUrl: string,
    private readonly token: string,
    private readonly organizationId?: string,
  ) {}

  /**
   * Establishes the Socket.IO connection. Resolves when connected.
   * Throws if the connection fails within the timeout.
   */
  async connect(timeoutMs = 5000): Promise<void> {
    if (this._connected && this.socket) {
      return
    }

    if (this.connectPromise) {
      return this.connectPromise
    }

    this.connectPromise = this.doConnect(timeoutMs)

    try {
      await this.connectPromise
    } finally {
      this.connectPromise = null
    }
  }

  private doConnect(timeoutMs: number): Promise<void> {
    return new Promise<void>((resolve, reject) => {
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

      const timer = setTimeout(() => {
        if (!this._connected) {
          this.socket?.disconnect()
          this._failed = true
          this._failError = 'WebSocket connection timed out'
          reject(new Error(this._failError))
        }
      }, timeoutMs)

      this.socket.on('connect', () => {
        clearTimeout(timer)
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
          clearTimeout(timer)
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

      // Register event listeners for all sandbox events
      this.socket.on('sandbox.state.updated', (data: SandboxStateUpdatedEvent) => {
        this.dispatch(data.sandbox?.id, { type: 'state.updated', data })
      })

      this.socket.on(
        'sandbox.desired-state.updated',
        (data: SandboxDesiredStateUpdatedEvent) => {
          this.dispatch(data.sandbox?.id, { type: 'desired-state.updated', data })
        },
      )

      this.socket.on('sandbox.created', (data: SandboxDto) => {
        this.dispatch(data.id, { type: 'created', data })
      })

      this.socket.connect()
    })
  }

  /**
   * Registers a handler for events targeting a specific sandbox.
   * Returns an unsubscribe function.
   */
  subscribe(sandboxId: string, handler: SandboxEventHandler): () => void {
    // Cancel any pending delayed disconnect since we have a new subscriber
    if (this.disconnectTimer) {
      clearTimeout(this.disconnectTimer)
      this.disconnectTimer = null
    }

    if (!this.listeners.has(sandboxId)) {
      this.listeners.set(sandboxId, new Set())
    }
    this.listeners.get(sandboxId)!.add(handler)

    return () => {
      const handlers = this.listeners.get(sandboxId)
      if (handlers) {
        handlers.delete(handler)
        if (handlers.size === 0) {
          this.listeners.delete(sandboxId)
        }
      }

      // Schedule delayed disconnect when no sandboxes are listening anymore
      if (this.listeners.size === 0) {
        this.disconnectTimer = setTimeout(() => {
          if (this.listeners.size === 0) {
            this.disconnect()
          }
        }, EventSubscriber.DISCONNECT_DELAY_MS)
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
    if (this.disconnectTimer) {
      clearTimeout(this.disconnectTimer)
      this.disconnectTimer = null
    }
    if (this.socket) {
      this.socket.removeAllListeners()
      this.socket.disconnect()
      this.socket = null
    }
    this._connected = false
    this.listeners.clear()
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

  private dispatch(sandboxId: string | undefined, event: SandboxEvent): void {
    if (!sandboxId) return
    const handlers = this.listeners.get(sandboxId)
    if (handlers) {
      for (const handler of handlers) {
        try {
          handler(event)
        } catch {
          // Don't let a handler error break other handlers
        }
      }
    }
  }
}
