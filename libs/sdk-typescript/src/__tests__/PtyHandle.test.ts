// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import type WebSocket from 'isomorphic-ws'
import { DaytonaConnectionError, DaytonaError, DaytonaTimeoutError } from '../errors/DaytonaError'

type WebSocketEventHandlers = {
  open?: () => void | Promise<void>
  message?: (event: { data: unknown } | unknown) => void | Promise<void>
  error?: (event: unknown) => void | Promise<void>
  close?: (event: { code?: number; reason?: string }) => void | Promise<void>
}

type MockWebSocket = {
  readyState: number
  binaryType?: string
  send: jest.Mock<void, [Uint8Array]>
  close: jest.Mock<void, []>
  addEventListener: jest.Mock<void, [string, (...args: never[]) => unknown]>
  handlers: WebSocketEventHandlers
}

const makeBrowserWebSocket = (): MockWebSocket => {
  const handlers: WebSocketEventHandlers = {}
  return {
    readyState: 0,
    binaryType: '',
    send: jest.fn(),
    close: jest.fn(),
    addEventListener: jest.fn((event: string, handler: (...args: never[]) => unknown) => {
      if (event === 'open') {
        handlers.open = handler as unknown as () => void | Promise<void>
      }
      if (event === 'message') {
        handlers.message = handler as unknown as (event: { data: unknown } | unknown) => void | Promise<void>
      }
      if (event === 'error') {
        handlers.error = handler as unknown as (event: unknown) => void | Promise<void>
      }
      if (event === 'close') {
        handlers.close = handler as unknown as (event: { code?: number; reason?: string }) => void | Promise<void>
      }
    }),
    handlers,
  }
}

describe('PtyHandle', () => {
  const makeHandle = async (wsOverrides?: Partial<MockWebSocket>) => {
    const { PtyHandle } = await import('../PtyHandle')
    const ws = Object.assign(makeBrowserWebSocket(), wsOverrides)
    const handleResize = jest.fn().mockResolvedValue({ sessionId: 'pty-1', cols: 120, rows: 40 })
    const handleKill = jest.fn().mockResolvedValue(undefined)
    const onPty = jest.fn().mockResolvedValue(undefined)

    const handle = new PtyHandle(ws as unknown as WebSocket, handleResize, handleKill, onPty, 'pty-1')

    return { handle, ws, handleResize, handleKill, onPty }
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('marks the socket as binary and disconnected by default', async () => {
    const { handle, ws } = await makeHandle()

    expect(ws.binaryType).toBe('arraybuffer')
    expect(handle.isConnected()).toBe(false)
    expect(handle.sessionId).toBe('pty-1')
  })

  it('waits for the server connected control message', async () => {
    jest.useFakeTimers()
    const { handle, ws } = await makeHandle()

    const waitPromise = handle.waitForConnection()
    ws.readyState = 1
    await ws.handlers.open?.()
    await ws.handlers.message?.({ data: JSON.stringify({ type: 'control', status: 'connected' }) })
    await jest.runOnlyPendingTimersAsync()

    await expect(waitPromise).resolves.toBeUndefined()
    expect(handle.isConnected()).toBe(true)
  })

  it('returns immediately when the connection is already established', async () => {
    const { handle, ws } = await makeHandle()

    ws.readyState = 1
    await ws.handlers.open?.()
    await ws.handlers.message?.({ data: JSON.stringify({ type: 'control', status: 'connected' }) })

    await expect(handle.waitForConnection()).resolves.toBeUndefined()
  })

  it('rejects waitForConnection when the server reports an error', async () => {
    jest.useFakeTimers()
    const { handle, ws } = await makeHandle()

    const waitPromise = handle.waitForConnection()
    const rejection = expect(waitPromise).rejects.toEqual(new DaytonaConnectionError('permission denied'))
    await ws.handlers.message?.({
      data: JSON.stringify({ type: 'control', status: 'error', error: 'permission denied' }),
    })
    await jest.runOnlyPendingTimersAsync()

    await rejection
    expect(handle.isConnected()).toBe(false)
  })

  it('rejects waitForConnection when the websocket closes before connecting', async () => {
    jest.useFakeTimers()
    const { handle } = await makeHandle({ readyState: 3 })

    const waitPromise = handle.waitForConnection()
    const rejection = expect(waitPromise).rejects.toEqual(new DaytonaConnectionError('Connection failed'))
    await jest.runOnlyPendingTimersAsync()

    await rejection
  })

  it('times out while waiting for connection', async () => {
    jest.useFakeTimers()
    const { handle } = await makeHandle()

    const waitPromise = handle.waitForConnection()
    const rejection = expect(waitPromise).rejects.toThrow('PTY connection timeout')
    await jest.advanceTimersByTimeAsync(10000)

    await rejection
    await expect(waitPromise).rejects.toBeInstanceOf(DaytonaTimeoutError)
  })

  it('sends string input as encoded bytes', async () => {
    const { handle, ws } = await makeHandle()

    ws.readyState = 1
    await ws.handlers.open?.()
    await ws.handlers.message?.({ data: JSON.stringify({ type: 'control', status: 'connected' }) })

    await handle.sendInput('ls -la\n')

    expect(ws.send).toHaveBeenCalledWith(new TextEncoder().encode('ls -la\n'))
  })

  it('sends raw binary input unchanged', async () => {
    const { handle, ws } = await makeHandle()
    const bytes = new Uint8Array([3])

    ws.readyState = 1
    await ws.handlers.open?.()
    await ws.handlers.message?.({ data: JSON.stringify({ type: 'control', status: 'connected' }) })

    await handle.sendInput(bytes)

    expect(ws.send).toHaveBeenCalledWith(bytes)
  })

  it('throws when sending input while disconnected', async () => {
    const { handle } = await makeHandle()

    await expect(handle.sendInput('pwd\n')).rejects.toBeInstanceOf(DaytonaConnectionError)
    await expect(handle.sendInput('pwd\n')).rejects.toThrow('PTY is not connected')
  })

  it('wraps send errors as DaytonaConnectionError', async () => {
    const { handle, ws } = await makeHandle()

    ws.readyState = 1
    ws.send.mockImplementation(() => {
      throw new Error('socket write failed')
    })
    await ws.handlers.open?.()
    await ws.handlers.message?.({ data: JSON.stringify({ type: 'control', status: 'connected' }) })

    await expect(handle.sendInput('pwd\n')).rejects.toBeInstanceOf(DaytonaConnectionError)
    await expect(handle.sendInput('pwd\n')).rejects.toThrow('Failed to send input to PTY: socket write failed')
  })

  it('delegates resize and kill operations', async () => {
    const { handle, handleResize, handleKill } = await makeHandle()

    await expect(handle.resize(120, 40)).resolves.toEqual({ sessionId: 'pty-1', cols: 120, rows: 40 })
    await handle.kill()

    expect(handleResize).toHaveBeenCalledWith(120, 40)
    expect(handleKill).toHaveBeenCalledTimes(1)
  })

  it('disconnects and ignores close errors', async () => {
    const { handle, ws } = await makeHandle()

    await handle.disconnect()
    expect(ws.close).toHaveBeenCalledTimes(1)

    ws.close.mockImplementation(() => {
      throw new Error('cannot close')
    })

    await expect(handle.disconnect()).resolves.toBeUndefined()
  })

  it('forwards regular text output to onPty', async () => {
    const { ws, onPty } = await makeHandle()

    await ws.handlers.message?.({ data: 'hello' })

    expect(onPty).toHaveBeenCalledWith(new TextEncoder().encode('hello'))
  })

  it('forwards binary ArrayBuffer output to onPty', async () => {
    const { ws, onPty } = await makeHandle()
    const bytes = new Uint8Array([104, 105])

    await ws.handlers.message?.({ data: bytes.buffer })

    expect(onPty).toHaveBeenCalledWith(new Uint8Array([104, 105]))
  })

  it('forwards Uint8Array views to onPty', async () => {
    const { ws, onPty } = await makeHandle()
    const bytes = new Uint8Array([1, 2, 3])

    await ws.handlers.message?.({ data: bytes })

    expect(onPty).toHaveBeenCalledWith(bytes)
  })

  it('captures websocket errors', async () => {
    const { handle, ws } = await makeHandle()

    await ws.handlers.error?.(new Error('boom'))

    await expect(handle.wait()).rejects.toBeInstanceOf(DaytonaError)
    await expect(handle.wait()).rejects.toThrow('boom')
    expect(handle.error).toBe('boom')
  })

  it('sets exit code to zero on normal close without a structured reason', async () => {
    const { handle, ws } = await makeHandle()

    await ws.handlers.close?.({ code: 1000, reason: '' })

    await expect(handle.wait()).resolves.toEqual({ exitCode: 0, error: undefined })
    expect(handle.exitCode).toBe(0)
  })

  it('parses structured close reasons with exit code and error details', async () => {
    const { handle, ws } = await makeHandle()

    await ws.handlers.close?.({
      code: 1006,
      reason: JSON.stringify({ exitCode: 17, exitReason: 'terminated', error: 'pty gone' }),
    })

    await expect(handle.wait()).resolves.toEqual({ exitCode: 17, error: 'pty gone' })
    expect(handle.exitCode).toBe(17)
    expect(handle.error).toBe('pty gone')
  })

  it('throws for unsupported websocket implementations', async () => {
    const { PtyHandle } = await import('../PtyHandle')
    const handleResize = jest.fn().mockResolvedValue({ sessionId: 'pty-1', cols: 80, rows: 24 })
    const handleKill = jest.fn().mockResolvedValue(undefined)
    const onPty = jest.fn().mockResolvedValue(undefined)

    expect(
      () =>
        new PtyHandle(
          { readyState: 0, send: jest.fn(), close: jest.fn() } as unknown as WebSocket,
          handleResize,
          handleKill,
          onPty,
          'pty-1',
        ),
    ).toThrow('Unsupported WebSocket implementation')
  })
})
