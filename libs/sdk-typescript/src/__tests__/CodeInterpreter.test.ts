// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import type { Configuration } from '@daytona/api-client'

const mockCreateSandboxWebSocket = jest.fn()

jest.mock('@daytona/toolbox-api-client', () => ({}), { virtual: true })
jest.mock('../utils/WebSocket', () => ({
  createSandboxWebSocket: (...args: unknown[]) => mockCreateSandboxWebSocket(...args),
}))

describe('CodeInterpreter', () => {
  const cfg: Configuration = {
    basePath: 'http://sandbox',
    baseOptions: { headers: { Authorization: 'Bearer token' } },
  } as unknown as Configuration

  const apiClient = {
    createInterpreterContext: jest.fn(),
    listInterpreterContexts: jest.fn(),
    deleteInterpreterContext: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  const makeInterpreter = async () => {
    const { CodeInterpreter } = await import('../CodeInterpreter')
    return new CodeInterpreter(cfg, apiClient as unknown as never, async () => 'preview')
  }

  it('validates code input', async () => {
    const interpreter = await makeInterpreter()
    await expect(interpreter.runCode('  ')).rejects.toThrow('Code is required for execution')
  })

  it('streams stdout/stderr/error chunks and resolves on completed control message', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const onStdout = jest.fn()
    const onStderr = jest.fn()
    const onError = jest.fn()

    const runPromise = interpreter.runCode('print(1)', { onStdout, onStderr, onError, timeout: 3 })
    await Promise.resolve()

    handlers.open()
    await handlers.message?.({ data: JSON.stringify({ type: 'stdout', text: 'ok-out' }) })
    await handlers.message?.({ data: JSON.stringify({ type: 'stderr', text: 'ok-err' }) })
    await handlers.message?.({
      data: JSON.stringify({ type: 'error', name: 'ValueError', value: 'bad', traceback: 'tb' }),
    })
    await handlers.message?.({ data: JSON.stringify({ type: 'control', text: 'completed' }) })

    const result = await runPromise
    expect(result.stdout).toBe('ok-out')
    expect(result.stderr).toBe('ok-err')
    expect(result.error).toEqual({ name: 'ValueError', value: 'bad', traceback: 'tb' })
    expect(onStdout).toHaveBeenCalled()
    expect(onStderr).toHaveBeenCalled()
    expect(onError).toHaveBeenCalled()
  })

  it('sends context, envs and timeout on websocket open', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const runPromise = interpreter.runCode('print(1)', {
      context: {
        id: 'ctx-1',
        active: true,
        createdAt: '2026-01-01T00:00:00Z',
        cwd: '/tmp',
        language: 'python',
      },
      envs: { PYTHONUNBUFFERED: '1' },
      timeout: 5,
    })
    await Promise.resolve()

    handlers.open()
    expect(ws.send).toHaveBeenCalledWith(
      JSON.stringify({
        code: 'print(1)',
        contextId: 'ctx-1',
        envs: { PYTHONUNBUFFERED: '1' },
        timeout: 5,
      }),
    )

    handlers.close?.({ code: 1000, reason: '' })
    await expect(runPromise).resolves.toEqual({ stdout: '', stderr: '' })
  })

  it('ignores invalid websocket payloads', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const runPromise = interpreter.runCode('print(1)')
    await Promise.resolve()

    handlers.open()
    await handlers.message?.({ data: 'not-json' })
    await handlers.message?.({ data: JSON.stringify({ type: 'unknown', text: 'ignored' }) })
    handlers.close?.({ code: 1000, reason: '' })

    await expect(runPromise).resolves.toEqual({ stdout: '', stderr: '' })
  })

  it('resolves on interrupted control messages', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const runPromise = interpreter.runCode('print(1)')
    await Promise.resolve()

    handlers.open()
    await handlers.message?.({ data: JSON.stringify({ type: 'control', text: 'interrupted' }) })

    await expect(runPromise).resolves.toEqual({ stdout: '', stderr: '' })
  })

  it('throws DaytonaTimeoutError when websocket closes with timeout code', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const runPromise = interpreter.runCode('print(1)', { timeout: 1 })
    await Promise.resolve()

    handlers.open()
    handlers.close?.({ code: 4008, reason: '' })

    await expect(runPromise).rejects.toThrow('Execution timed out')
  })

  it('maps websocket close reasons to DaytonaConnectionError', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const runPromise = interpreter.runCode('print(1)')
    await Promise.resolve()

    handlers.open()
    handlers.close?.({ code: 1011, reason: 'runner crashed' })

    await expect(runPromise).rejects.toThrow('runner crashed (close code 1011)')
  })

  it('maps websocket error events to DaytonaConnectionError', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const runPromise = interpreter.runCode('print(1)')
    await Promise.resolve()

    handlers.error?.(new Error('socket exploded'))

    await expect(runPromise).rejects.toThrow('Failed to execute code: socket exploded')
  })

  it('supports node-style websocket implementations with on/off methods', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      on: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      off: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const runPromise = interpreter.runCode('print(1)')
    await Promise.resolve()

    handlers.open()
    handlers.close?.({ code: 1000, reason: '' })

    await expect(runPromise).resolves.toEqual({ stdout: '', stderr: '' })
    expect(ws.on).toHaveBeenCalled()
  })

  it('returns an empty list when no interpreter contexts exist', async () => {
    const interpreter = await makeInterpreter()
    apiClient.listInterpreterContexts.mockResolvedValue({ data: {} })

    await expect(interpreter.listContexts()).resolves.toEqual([])
  })

  it('delegates context CRUD methods', async () => {
    const interpreter = await makeInterpreter()
    apiClient.createInterpreterContext.mockResolvedValue({ data: { id: 'ctx-1', cwd: '/tmp' } })
    apiClient.listInterpreterContexts.mockResolvedValue({ data: { contexts: [{ id: 'ctx-1' }] } })
    apiClient.deleteInterpreterContext.mockResolvedValue({ data: undefined })

    const ctx = await interpreter.createContext('/tmp')
    expect(ctx).toEqual({ id: 'ctx-1', cwd: '/tmp' })
    await expect(interpreter.listContexts()).resolves.toEqual([{ id: 'ctx-1' }])
    await interpreter.deleteContext({
      id: 'ctx-1',
      active: true,
      createdAt: '2026-01-01T00:00:00Z',
      cwd: '/tmp',
      language: 'python',
    })
  })

  it('supports extractMessageText helpers for raw payload types', async () => {
    const interpreter = await makeInterpreter()
    const runtime = interpreter as unknown as {
      extractMessageText: (event: unknown) => Promise<string>
    }

    await expect(runtime.extractMessageText(new TextEncoder().encode('abc'))).resolves.toBe('abc')
    await expect(runtime.extractMessageText(new TextEncoder().encode('abc').buffer)).resolves.toBe('abc')
    await expect(runtime.extractMessageText({ data: null })).resolves.toBe('')
  })
})
